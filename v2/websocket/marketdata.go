package websocket

import (
	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"context"
)

type MarketData interface {
	SubscribeTicker(symbol string) <-chan *bitfinex.Ticker
	SubscribeTrades(symbol string) <-chan *bitfinex.Trade
}

func (c Client) SubscribeTicker(ctx context.Context, symbol string) (<-chan *bitfinex.Ticker, error) {
	ch := make(chan *bitfinex.Ticker)
	req := &subscriptionRequest{
		SubID: c.subscriptions.NextSubID(),
		Event: EventSubscribe,
		Channel: ChanTicker,
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
			// already strongly typed
			tick, ok := m.(*bitfinex.Ticker)
			if ok {
				ch <- tick
			}
		}
	}()
	return ch, nil
}

func (c Client) SubscribeTrades(ctx context.Context, symbol string) (<-chan *bitfinex.Trade, error) {
	ch := make(chan *bitfinex.Trade)
	req := &subscriptionRequest{
		SubID: c.subscriptions.NextSubID(),
		Event: EventSubscribe,
		Channel: ChanTrades,
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
			// already strongly typed
			trade, ok := m.(*bitfinex.Trade)
			if ok {
				ch <- trade
			}
		}
	}()
	return ch, nil
}