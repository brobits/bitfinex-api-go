package websocket

import (
	"context"
	"fmt"

	"github.com/bitfinexcom/bitfinex-api-go/utils"
	"github.com/bitfinexcom/bitfinex-api-go/v2/domain"
	"golang.org/x/tools/go/gcimporter15/testdata"
)

type unsubscribeMsg struct {
	Event  string `json:"event"`
	ChanID int64  `json:"chanId"`
}

// Unsubscribe from the websocket channel with the given channel id and close
// the associated go channel.
func (c Client) UnsubscribeByChanID(ctx context.Context, id int64) error {
	err := c.Subscriptions.RemoveByChanID(id)
	if err != nil {
		return err
	}
	return c.sendUnsubscribeMessage(ctx, id)
}

// Unsubscribe takes an PublicSubscriptionRequest and tries to unsubscribe from the
// channel described by that request.
func (c Client) Unsubscribe(ctx context.Context, p *PublicSubscriptionRequest) error {
	if p == nil {
		return fmt.Errorf("PublicSubscriptionRequest cannot be nil")
	}

	for k, v := range b.pubChanIDs {
		if v == *p {
			return c.UnsubscribeByChanID(ctx, k)
		}
	}
	return fmt.Errorf("could not find channel for symbol")
}

func (c Client) sendUnsubscribeMessage(ctx context.Context, id int64) error {
	return c.Send(ctx, unsubscribeMsg{Event: "unsubscribe", ChanID: id})
}

// PublicSubscriptionRequest is used to subscribe to one of the public websocket
// channels. The `Event` field is automatically set to `subscribe` when using the
// Subscribe method. The `Channel` field is mandatory. For all other fields please
// consult the officical documentation here: http://docs.bitfinex.com/v2/reference#ws-public-ticker
type PublicSubscriptionRequest struct {
	Event     string `json:"event"`
	Channel   string `json:"channel"`
	Symbol    string `json:"symbol"`
	Precision string `json:"prec,omitempty"`
	Frequency string `json:"freq,omitempty"`
	Key       string `json:"key,omitempty"`
	Len       string `json:"len,omitempty"`
	Pair      string `json:"pair,omitempty"`
	SubID     string `json:"subId,omitempty"`
}

// Subscribe to one of the public websocket channels.
func (c Client) Subscribe(ctx context.Context, msg *PublicSubscriptionRequest, h handlerT) error {
	if c.ws == nil {
		return ErrWSNotConnected
	} else if msg == nil {
		return fmt.Errorf("no subscription request provided")
	}

	for _, v := range b.pubChanIDs {
		if v == *msg {
			return fmt.Errorf("already subscribed to the channel requested")
		}
	}

	msg.Event = "subscribe"
	msg.SubID = utils.GetNonce()

	b.subMu.Lock()
	b.pubSubIDs[msg.SubID] = publicSubInfo{req: *msg, h: h}
	b.subMu.Unlock()

	return b.Send(ctx, msg)
}

func (c Client) handlePublicDataMessage(raw []interface{}) (interface{}, error) {
	switch len(raw) {
	case 2:
		// [ChanID, [Data]] or [ChanID, "hb"]
		// Data can be either []float64 or [][]float64, where the former should be
		// representing an update and the latter a snapshot.
		// Simple update/snapshot for ticker, books, raw books and candles.
		switch fp := raw[1].(type) {
		case []interface{}:
			return c.processDataSlice(fp)
		case string: // This should be a heartbeat.
			return domain.Heartbeat{}, nil
		}
	case 3:
		// [ChanID, MsgType, [Data]]
		// Data can be either []float64 or [][]float64, where the former should be
		// representing an update and the latter a snapshot.
		if fp, ok := raw[2].([]interface{}); ok {
			return c.processDataSlice(fp)
		}
	}

	return nil, fmt.Errorf("unexpected data message: %#v", raw)
}

func (c Client) processDataSlice(data []interface{}) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("unexpected data slice: %v", data)
	}

	var items [][]float64
	switch data[0].(type) {
	case []interface{}: // [][]float64
		for _, e := range data {
			if s, ok := e.([]interface{}); ok {
				item, err := domain.F64Slice(s)
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			} else {
				return nil, fmt.Errorf("expected slice of float64 slices but got: %v", data)
			}
		}
	case float64: // []float64
		item, err := domain.F64Slice(data)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	default:
		return nil, fmt.Errorf("unexpected data slice: %v", data)
	}

	return items, nil
}
