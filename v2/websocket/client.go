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
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"github.com/bitfinexcom/bitfinex-api-go/v2"
)

var productionBaseURL = "wss://api.bitfinex.com/ws/2"

// Available channels
const (
	ChanBook    = "book"
	ChanTrades  = "trades"
	ChanTicker  = "ticker"
	ChanCandles = "candles"
)

// Events
const (
	EventSubscribe 		= "subscribe"
	EventUnsubscribe 	= "unsubscribe"
)

var (
	ErrWSNotConnected     = fmt.Errorf("websocket connection not established")
	ErrWSAlreadyConnected = fmt.Errorf("websocket connection already established")
)

type authState byte
type AuthState authState // prevent user construction of authStates

const (
	NoAuthentication AuthState = 0
	PendingAuthentication AuthState = 1
	SuccessfulAuthentication AuthState = 2
	RejectedAuthentication AuthState = 3
)

type Asynchronous interface {
	Send(ctx context.Context, msg interface{}) error
	Listen(chanID int64) <-chan []interface{}
}

type Client struct {
	ws				*websocket.Conn
	wsLock			sync.Mutex
	BaseURL 		string
	TLSSkipVerify 	bool
	timeout			int64 // read timeout
	apiKey			string
	apiSecret		string
	Authentication	AuthState

	// subscription manager
	subscriptions 	Subscriptions
	Asynchronous
	factories		map[string]messageFactory

	// websocket shutdown signal
	shutdown		chan error

	// close signal sent to user on shutdown.  ws -> shutdown channel -> internal cleanup -> done channel
	done			chan error

	// event forwarding
	eventListener EventListener
}

func (c Client) AttachEventListener(listener EventListener) {
	c.eventListener = listener
}

func (c Client) Credentials(key string, secret string) *Client {
	c.apiKey = key
	c.apiSecret = secret
	return &c
}

func (c Client) sign(msg string) string {
	sig := hmac.New(sha512.New384, []byte(c.apiSecret))
	sig.Write([]byte(msg))
	return hex.EncodeToString(sig.Sum(nil))
}

func (c Client) registerFactory(channel string, factory messageFactory) {
	c.factories[channel] = factory
}

func NewClientWithUrl(url string) *Client {
	c := &Client{
		BaseURL: url,
		shutdown: make(chan error),
		done: make(chan error),
		Authentication: NoAuthentication,
		factories: make(map[string]messageFactory),
	}
	c.registerFactory(ChanTicker, func(raw []interface{}) (msg interface{}, err error) {
		return bitfinex.NewTickerFromRaw(raw)
	})
	return c
}

func NewClient() *Client {
	return NewClientWithUrl(productionBaseURL)
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

func (c Client) listen(subID string) (<-chan interface{}, error) {
	sub, err := c.subscriptions.LookupBySubscriptionID(subID)
	if err != nil {
		return nil, err
	}
	return sub.Stream(), nil
}

func (c Client) handleMessage(msg []byte) error {
	t := bytes.TrimLeftFunc(msg, unicode.IsSpace)
	err := error(nil)
	// either a channel data array or an event object, raw json encoding
	if bytes.HasPrefix(t, []byte("[")) {
		err = c.handleChannel(msg)
	} else if bytes.HasPrefix(t, []byte("{")) {
		err = c.handleEvent(msg)
	} else {
		return fmt.Errorf("unexpected message: %s", msg)
	}
	return err
}

// Unsubscribe takes an PublicSubscriptionRequest and tries to unsubscribe from the
// channel described by that request.
/*
func (c Client) Unsubscribe(ctx context.Context, p *PublicSubscriptionRequest) error {
	if p == nil {
		return fmt.Errorf("PublicSubscriptionRequest cannot be nil")
	}
	c.subscriptions.RemoveBySubID(p.SubID)
	return fmt.Errorf("could not find channel for symbol")
}
*/

func (c Client) sendUnsubscribeMessage(ctx context.Context, id int64) error {
	return c.Send(ctx, unsubscribeMsg{Event: "unsubscribe", ChanID: id})
}

/*
// Subscribe to one of the public websocket channels.
func (c Client) Subscribe(ctx context.Context, msg *PublicSubscriptionRequest) (<-chan []interface{}, error) {
	if c.ws == nil {
		return nil, ErrWSNotConnected
	} else if msg == nil {
		return nil, fmt.Errorf("no subscription request provided")
	}

	msg.Event = "subscribe"
	if msg.SubID == "" {
		msg.SubID = utils.GetNonce()
	}

	if _, err := c.subscriptions.LookupBySubscriptionID(msg.SubID); err == nil {
		return nil, fmt.Errorf("subscription exists for sub ID %s", msg.SubID)
	}

	sub := c.subscriptions.Add(msg)
	err := c.Send(ctx, msg)
	if err != nil {
		c.subscriptions.RemoveBySubID(msg.SubID)
		return nil, err
	}
	return sub.Stream(), nil
}
*/

// Unsubscribe from the websocket channel with the given channel id and close
// the associated go channel.
func (c Client) UnsubscribeByChanID(ctx context.Context, id int64) error {
	err := c.subscriptions.RemoveByChanID(id)
	if err != nil {
		return err
	}
	return c.sendUnsubscribeMessage(ctx, id)
}

// Authenticate creates the payload for the authentication request and sends it
// to the API. The filters will be applied to the authenticated channel, i.e.
// only subscribe to the filtered messages.
func (c Client) Authenticate(ctx context.Context, filter ...string) error {
	nonce := utils.GetNonce()

	payload := "AUTH" + nonce
	s := &subscriptionRequest{
		Event:       "auth",
		APIKey:      c.apiKey,
		AuthSig:     c.sign(payload),
		AuthPayload: payload,
		AuthNonce:   nonce,
		Filter:      filter,
		SubID:       nonce,
	}
	c.subscriptions.Add(s)

	if err := c.Send(ctx, s); err != nil {
		return err
	}
	c.Authentication = PendingAuthentication

	return nil
}

// SetReadTimeout sets the read timeout for the underlying websocket connections.
func (c Client) SetReadTimeout(t time.Duration) {
	atomic.StoreInt64(&c.timeout, t.Nanoseconds())
}

// TODO auto subscription with connect
/*
	async := e.Asynchronous.Listen()
	go func() {
		for o := range async {
			if o == nil {
				// channel closed, propagate EOT
				close(ch)
			}
			tick, err := bitfinex.NewTickerFromRaw(o)
			if err != nil {
				log.Printf("could not crack message: %s", err.Error())
				continue
			}
			ch <- &tick
		}
	}()
 */
/*
func ExampleUsage() {
	client := ExampleClient{}
	ch, err := client.SubscribeTicker("tBTCUSD")
	if err != nil {
		// error subscribing
		return
	}
	for tick := range ch {
		if tick == nil {
			// channel closed
			return
		}
		// TODO: handle tick
	}
}
*/