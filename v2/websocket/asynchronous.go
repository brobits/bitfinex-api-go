package websocket

import (
	"github.com/bitfinexcom/bitfinex-api-go/v2/domain"
	"context"
)

type Asynchronous interface {
	Send(ctx context.Context, msg interface{}) error
	Listen(chanID int64) <-chan []interface{}
}

type MarketData interface {
	SubscribeTicker(symbol string) <-chan *domain.Ticker
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
			tick, err := domain.NewTickerFromRaw(o)
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