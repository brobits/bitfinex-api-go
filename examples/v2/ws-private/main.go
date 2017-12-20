package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/bitfinexcom/bitfinex-api-go/v2/websocket"
)

// Set BFX_APIKEY and BFX_SECRET as :
//
// export BFX_API_KEY=YOUR_API_KEY
// export BFX_API_SECRET=YOUR_API_SECRET
//
// you can obtain it from https://www.bitfinex.com/api

type TestEventListener struct {
	websocket.DefaultEventListener
}

func (t TestEventListener) onRawEvent(ev interface{}) {
	log.Printf("EVENT: %#v", ev)
}

func main() {

	key := os.Getenv("BFX_API_KEY")
	secret := os.Getenv("BFX_API_SECRET")
	c := websocket.NewClient().Credentials(key, secret)

	err := c.Connect()
	if err != nil {
		log.Fatalf("connecting authenticated websocket: %s", err)
	}

	c.AttachEventListener(&TestEventListener{})
/*
	c.AttachPrivateHandler(func(msg interface{}) {
		log.Printf("PRIV MSG: %#v", msg)
	})
*/
	ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
	c.Authenticate(ctx)

	time.Sleep(time.Second * 10)
}
