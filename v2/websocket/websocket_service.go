package websocket

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/bitfinexcom/bitfinex-api-go/utils"

	"github.com/gorilla/websocket"
	"golang.org/x/tools/go/gcimporter15/testdata"
	"github.com/bitfinexcom/bitfinex-api-go/v2/domain"
)

// Available channels
const (
	ChanBook    = "book"
	ChanTrades  = "trades"
	ChanTicker  = "ticker"
	ChanCandles = "candles"
)

var (
	ErrWSNotConnected     = fmt.Errorf("websocket connection not established")
	ErrWSAlreadyConnected = fmt.Errorf("websocket connection already established")
)

type Client struct {
	ws				*websocket.Conn
	wsLock			sync.Mutex
	BaseURL 		string
	TLSSkipVerify 	bool
	isAuthenticated	bool
	timeout			int64 // read timeout

	// subscription manager
	subscriptions Subscriptions
	Asynchronous

	// websocket shutdown signal
	shutdown		chan error

	// close signal sent to user on shutdown.  ws -> shutdown channel -> internal cleanup -> done channel
	done			chan error
}

func NewClient(url string) *Client {
	return &Client{
		BaseURL: url,
		shutdown: make(chan error),
		done: make(chan error),
	}
}

func (c Client) listenWs() {
	for {
		if c.ws == nil {
			return
		}
		if atomic.LoadInt64(&c.timeout) != 0 {
			c.ws.SetReadDeadline(time.Now().Add(time.Duration(c.timeout)))
		}

		select {
		case err := <-c.shutdown:
			// websocket termination
			if err != nil {
				c.done <- err
			}
			close(c.done)
			return
		default:
		}

		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			c.close(err)
			return
		}

		// Errors here should be non critical so we just log them.
		err = c.handleMessage(msg)
		if err != nil {
			log.Printf("[WARN]: %s\n", err)
		}
	}
}

func (c Client) Connect() error {
	if c.ws != nil {
		return nil // no op
	}
	// init?
	return c.connect()
}

func (c Client) connect() error {
	c.wsLock.Lock()
	defer c.wsLock.Unlock()
	var d = websocket.Dialer{
		Subprotocols:    []string{"p1", "p2"},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		Proxy:           http.ProxyFromEnvironment,
	}

	d.TLSClientConfig = &tls.Config{InsecureSkipVerify: c.TLSSkipVerify}

	ws, _, err := d.Dial(c.BaseURL, nil)
	if err != nil {
		return err
	}

	c.ws = ws
	go c.listenWs()
	return nil
}

// Done returns a channel that will be closed if the underlying websocket
// connection gets closed.
func (c Client) Done() <-chan error { return c.done }

func (c Client) close(e error) {
	c.wsLock.Lock()
	if c.ws != nil {
		if err := c.ws.Close(); err != nil {
			log.Printf("[INFO]: error closing websocket: %s", err)
		}
		c.ws = nil
	}
	c.wsLock.Unlock()

	// send error to shutdown channel
	c.shutdown <- e

	// close channel
	close(c.shutdown)
}

func (c Client) Close() {
	c.close(nil)
}

// Send marshals the given interface and then sends it to the API. This method
// can block so specify a context with timeout if you don't want to wait for too
// long.
func (c Client) Send(ctx context.Context, msg interface{}) error {
	if c.ws == nil {
		return ErrWSNotConnected
	}

	bs, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.Done():
		return fmt.Errorf("websocket closed: ", "error msg") // TODO pass error msg
	default:
	}

	c.wsLock.Lock()
	defer c.wsLock.Unlock()
	err = c.ws.WriteMessage(websocket.TextMessage, bs)
	if err != nil {
		c.close(err)
		return err
	}

	return nil
}

func (c Client) listen(subID string) (<-chan []interface{}, error) {
	sub, err := c.subscriptions.LookupBySubscriptionID(subID)
	if err != nil {
		return nil, err
	}
	return sub.Stream(), nil
}

func (c Client) SubscribeTicker(ctx context.Context, symbol string) (<-chan *domain.Ticker, error) {
	ch := make(chan *domain.Ticker)
	req := &PublicSubscriptionRequest{
		SubID: c.subscriptions.NextSubID(),
		Event: "subscribe",
		Channel: "ticker",
		Symbol: symbol,
	}
	err := c.Asynchronous.Send(ctx, req)
	if err != nil {
		// propagate send error
		return nil, err
	}
	pipe, err := c.listen(req.SubID)
	if err != nil {
		// propagate no sub error
		return nil, err
	}
	go func() {
		for m := range pipe {
			if m == nil {
				// channel closed, propagate EOT
				close(ch)
				break
			}
			tick, err := domain.NewTickerFromRaw(m)
			if err != nil {
				log.Printf("could not convert ticker message: %s", err.Error())
				continue
			}
			ch <- &tick
		}
	}()
	return ch, nil
}

func (c Client) handleMessage(msg []byte) error {
	t := bytes.TrimLeftFunc(msg, unicode.IsSpace)
	if bytes.HasPrefix(t, []byte("[")) { // Data messages are just arrays of values.
		var raw []interface{}
		err := json.Unmarshal(msg, &raw)
		if err != nil {
			return err
		} else if len(raw) < 2 {
			return nil
		}

		chanID, ok := raw[0].(float64)
		if !ok {
			return fmt.Errorf("expected message to start with a channel id but got %#v instead", raw[0])
		}

		if _, has := b.privChanIDs[int64(chanID)]; has {
			td, err := b.handlePrivateDataMessage(raw)
			if err != nil {
				return err
			} else if td == nil {
				return nil
			}
			if b.privateHandler != nil {
				go b.privateHandler(td)
				return nil
			}
		} else if _, has := b.pubChanIDs[int64(chanID)]; has {
			td, err := b.handlePublicDataMessage(raw)
			if err != nil {
				return err
			} else if td == nil {
				return nil
			}
			if h, has := b.publicHandlers[int64(chanID)]; has {
				go h(td)
				return nil
			}
		} else {
			// TODO: log unhandled message?
		}
	} else if bytes.HasPrefix(t, []byte("{")) { // Events are encoded as objects.
		ev, err := b.onEvent(msg)
		if err != nil {
			return err
		}
		if b.eventHandler != nil {
			go b.eventHandler(ev)
		}
	} else {
		return fmt.Errorf("unexpected message: %s", msg)
	}

	return nil
}

type subscriptionRequest struct {
	Event       string   `json:"event"`
	APIKey      string   `json:"apiKey"`
	AuthSig     string   `json:"authSig"`
	AuthPayload string   `json:"authPayload"`
	AuthNonce   string   `json:"authNonce"`
	Filter      []string `json:"filter"`
	SubID       string   `json:"subId"`
}

// Authenticate creates the payload for the authentication request and sends it
// to the API. The filters will be applied to the authenticated channel, i.e.
// only subscribe to the filtered messages.
func (b *bfxWebsocket) Authenticate(ctx context.Context, filter ...string) error {
	nonce := utils.GetNonce()
	payload := "AUTH" + nonce
	s := &subscriptionRequest{
		Event:       "auth",
		APIKey:      b.client.APIKey,
		AuthSig:     b.client.sign(payload),
		AuthPayload: payload,
		AuthNonce:   nonce,
		Filter:      filter,
		SubID:       nonce,
	}

	b.subMu.Lock()
	b.privSubIDs[nonce] = struct{}{}
	b.subMu.Unlock()

	if err := b.Send(ctx, s); err != nil {
		return err
	}
	b.isAuthenticated = true

	return nil
}

func (b *bfxWebsocket) AttachEventHandler(f handlerT) error {
	b.eventHandler = f
	return nil
}

func (b *bfxWebsocket) AttachPrivateHandler(f handlerT) error {
	b.privateHandler = f
	return nil
}

func (b *bfxWebsocket) RemoveEventHandler() error {
	b.eventHandler = nil
	return nil
}

func (b *bfxWebsocket) RemovePrivateHandler() error {
	b.privateHandler = nil
	return nil
}

// SetReadTimeout sets the read timeout for the underlying websocket connections.
func (c Client) SetReadTimeout(t time.Duration) {
	atomic.StoreInt64(&c.timeout, t.Nanoseconds())
}
