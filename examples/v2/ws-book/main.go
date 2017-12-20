package main

import (
	"context"
	"log"
	"time"

	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/bitfinexcom/bitfinex-api-go/v2/websocket"
)

func main() {
	c := websocket.NewClient()

	err := c.Connect()
	if err != nil {
		log.Fatal("Error connecting to web socket : ", err)
	}
	c.SetReadTimeout(time.Second * 2)

	// subscribe to BTCUSD ticker
	ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
	tkr, err := c.SubscribeTicker(ctx, bitfinex.TradingPrefix + bitfinex.BTCUSD)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for obj := range tkr {
			if obj == nil {
				break
			}
			log.Printf("PUBLIC MSG BTCUSD: %#v", obj)
		}
	}()

	// subscribe to IOTUSD trades
	ctx, _ = context.WithTimeout(context.Background(), time.Second*1)
	tds, err := c.SubscribeTrades(ctx, bitfinex.TradingPrefix + bitfinex.IOTUSD)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for obj := range tds {
			if obj == nil {
				break
			}
			log.Printf("PUBLIC MSG IOTUSD: %#v", obj)
		}
	}()

	// listen for ws disconnect
	for {
		select {
		case m := <-c.Done():
			if m != nil {
				log.Printf("channel closed: %s", m)
			}
			return
		}
	}
}
