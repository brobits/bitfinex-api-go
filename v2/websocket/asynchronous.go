package websocket

import (
	"log"
)

type Asynchronous interface {
	Send(request interface{}) (<-chan interface{}, error)
}

// AsyncCracker rather than generic Cracker to encapsulate WebSocket-specific domain messages
type AsyncCracker interface {
	CrackTicker(msg interface{}) (*Ticker, error)
}

type MarketData interface {
	SubscribeTicker(symbol string) <-chan *Ticker
}

type ExampleClient struct {
	Asynchronous
	AsyncCracker
}

func (e ExampleClient) CrackTicker(msg interface{}) (*Ticker, error) {
	// TODO use websocket service to crack interface into strongly-typed ticker w/ error response
	return &Ticker{}, nil
}

func (e ExampleClient) SubscribeTicker(symbol string) (<-chan *Ticker, error) {
	ch := make(chan *Ticker)
	req := &PublicSubscriptionRequest{
		Event: "subscribe",
		Channel: "ticker",
		Symbol: symbol,
	}
	async, err := e.Asynchronous.Send(req)
	if err != nil {
		// propagate error
		return nil, err
	}
	go func() {
		for o := range async {
			if o == nil {
				// channel closed, propagate EOT
				close(ch)
			}
			tick, err := e.AsyncCracker.CrackTicker(o)
			if err != nil {
				log.Printf("could not crack message: %s", err.Error())
				continue
			}
			ch <- tick
		}
	}()
	return ch, nil
}

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