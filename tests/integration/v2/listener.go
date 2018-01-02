package tests

import (
	"errors"
	"log"
	"time"

	bitfinex "github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/bitfinexcom/bitfinex-api-go/v2/websocket"
)

type listener struct {
	infoEvents           chan *websocket.InfoEvent
	authEvents           chan *websocket.AuthEvent
	ticks                chan *bitfinex.Ticker
	subscriptionEvents   chan *websocket.SubscribeEvent
	unsubscriptionEvents chan *websocket.UnsubscribeEvent
	errors               chan error
}

func newListener() *listener {
	return &listener{
		infoEvents:           make(chan *websocket.InfoEvent, 10),
		authEvents:           make(chan *websocket.AuthEvent, 10),
		ticks:                make(chan *bitfinex.Ticker, 10),
		subscriptionEvents:   make(chan *websocket.SubscribeEvent, 10),
		unsubscriptionEvents: make(chan *websocket.UnsubscribeEvent, 10),
		errors:               make(chan error, 10),
	}
}

func (l *listener) nextInfoEvent() (*websocket.InfoEvent, error) {
	timeout := make(chan bool)
	go func() {
		time.Sleep(time.Second * 2)
		close(timeout)
	}()
	select {
	case ev := <-l.infoEvents:
		return ev, nil
	case <-timeout:
		return nil, errors.New("timed out waiting for InfoEvent")
	}
}

func (l *listener) nextAuthEvent() (*websocket.AuthEvent, error) {
	timeout := make(chan bool)
	go func() {
		time.Sleep(time.Second * 2)
		close(timeout)
	}()
	select {
	case ev := <-l.authEvents:
		return ev, nil
	case <-timeout:
		return nil, errors.New("timed out waiting for AuthEvent")
	}
}

func (l *listener) nextSubscriptionEvent() (*websocket.SubscribeEvent, error) {
	timeout := make(chan bool)
	go func() {
		time.Sleep(time.Second * 2)
		close(timeout)
	}()
	select {
	case ev := <-l.subscriptionEvents:
		return ev, nil
	case <-timeout:
		return nil, errors.New("timed out waiting for SubscribeEvent")
	}
}

func (l *listener) nextUnsubscriptionEvent() (*websocket.UnsubscribeEvent, error) {
	timeout := make(chan bool)
	go func() {
		time.Sleep(time.Second * 2)
		close(timeout)
	}()
	select {
	case ev := <-l.unsubscriptionEvents:
		return ev, nil
	case <-timeout:
		return nil, errors.New("timed out waiting for UnsubscribeEvent")
	}
}

func (l *listener) nextTick() (*bitfinex.Ticker, error) {
	timeout := make(chan bool)
	go func() {
		time.Sleep(time.Second * 2)
		close(timeout)
	}()
	select {
	case ev := <-l.ticks:
		return ev, nil
	case <-timeout:
		return nil, errors.New("timed out waiting for Ticker")
	}
}

// strongly types messages and places them into a channel
func (l *listener) run(ch <-chan interface{}) {
	go func() {
		for {
			select {
			case msg := <-ch:
				if msg == nil {
					return
				}
				// remove threading guarantees when mulitplexing into channels
				log.Printf("%#v", msg)
				switch msg.(type) {
				case error:
					l.errors <- msg.(error)
				case *bitfinex.Ticker:
					l.ticks <- msg.(*bitfinex.Ticker)
				case *websocket.InfoEvent:
					l.infoEvents <- msg.(*websocket.InfoEvent)
				case *websocket.SubscribeEvent:
					l.subscriptionEvents <- msg.(*websocket.SubscribeEvent)
				case *websocket.UnsubscribeEvent:
					l.unsubscriptionEvents <- msg.(*websocket.UnsubscribeEvent)
				case *websocket.AuthEvent:
					l.authEvents <- msg.(*websocket.AuthEvent)
				}
			}
		}
	}()
}
