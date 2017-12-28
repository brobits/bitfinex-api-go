package tests

import (
	"context"
	"testing"
	"time"

	bitfinex "github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/bitfinexcom/bitfinex-api-go/v2/websocket"
)

func TestTicker(t *testing.T) {
	// test setup
	rawInfoEvent := []byte{123, 34, 101, 118, 101, 110, 116, 34, 58, 34, 105, 110, 102, 111, 34, 44, 34, 118, 101, 114, 115, 105, 111, 110, 34, 58, 50, 125}
	rawSubAck := []byte{123, 34, 101, 118, 101, 110, 116, 34, 58, 34, 115, 117, 98, 115, 99, 114, 105, 98, 101, 100, 34, 44, 34, 99, 104, 97, 110, 110, 101, 108, 34, 58, 34, 116, 105, 99, 107, 101, 114, 34, 44, 34, 99, 104, 97, 110, 73, 100, 34, 58, 53, 44, 34, 115, 121, 109, 98, 111, 108, 34, 58, 34, 116, 66, 84, 67, 85, 83, 68, 34, 44, 34, 115, 117, 98, 73, 100, 34, 58, 34, 49, 53, 49, 52, 52, 48, 49, 49, 55, 51, 48, 48, 49, 34, 44, 34, 112, 97, 105, 114, 34, 58, 34, 66, 84, 67, 85, 83, 68, 34, 125}
	rawTick := []byte{91, 53, 44, 91, 49, 52, 57, 53, 55, 44, 54, 56, 46, 49, 55, 51, 50, 56, 55, 57, 54, 44, 49, 52, 57, 53, 56, 44, 53, 53, 46, 50, 57, 53, 56, 56, 49, 51, 50, 44, 45, 54, 53, 57, 44, 45, 48, 46, 48, 52, 50, 50, 44, 49, 52, 57, 55, 49, 44, 53, 51, 55, 50, 51, 46, 48, 56, 56, 49, 51, 57, 57, 53, 44, 49, 54, 52, 57, 52, 44, 49, 52, 52, 53, 52, 93, 93}

	// create transport & nonce mocks
	async := newTestAsync()
	nonce := &MockNonceGenerator{}

	// create client
	ws := websocket.NewClientWithAsyncNonce(async, nonce)

	// setup listener
	listener := newListener()
	listener.run(ws.Listen())

	// set ws options
	ws.SetReadTimeout(time.Second * 2)
	ws.Connect()
	defer ws.Close()

	// begin test
	async.Publish(rawInfoEvent)
	nonce.Next("1514401173001")
	ev, err := listener.nextInfoEvent()
	if err != nil {
		t.Fatal(err)
	}
	assert(t, &websocket.InfoEvent{}, ev)

	// subscribe
	id, err := ws.SubscribeTicker(context.Background(), "tBTCUSD")
	if err != nil {
		t.Fatal(err)
	}
	async.Publish(rawSubAck)
	sub, err := listener.nextSubscriptionEvent()
	if err != nil {
		t.Fatal(err)
	}
	assert(t, &websocket.SubscribeEvent{
		SubID:   "1514401173001",
		Channel: "ticker",
		ChanID:  5,
		Symbol:  "tBTCUSD",
		Pair:    "BTCUSD",
	}, sub)

	async.Publish(rawTick)
	tick, err := listener.nextTick()
	if err != nil {
		t.Fatal(err)
	}
	assert(t, &bitfinex.Ticker{
		Symbol:          "tBTCUSD",
		Bid:             14957,
		Ask:             14958,
		BidSize:         68.17328796,
		AskSize:         55.29588132,
		DailyChange:     -659,
		DailyChangePerc: -0.0422,
		LastPrice:       14971,
		Volume:          53723.08813995,
		High:            16494,
		Low:             14454,
	}, tick)

	// unsubscribe
	ws.Unsubscribe(context.Background(), id)
	/*
		TODO
		async.Publish(rawUnsubAck)
		unsub, err := listener.nextUnsubscriptionEvent()
		if err != nil {
			t.Fatal(err)
		}
		assert(t, &websocket.UnsubscribeEvent{}, unsub)
	*/
}
