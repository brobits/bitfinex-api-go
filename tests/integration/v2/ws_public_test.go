package tests

import (
	"context"
	"log"
	"reflect"
	"testing"
	"time"

	bitfinex "github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/bitfinexcom/bitfinex-api-go/v2/websocket"
)

func assertTicker(t *testing.T, expected *bitfinex.Ticker, actual *bitfinex.Ticker) {
	if expected.Symbol != "" && expected.Symbol != actual.Symbol {
		t.Errorf("expected symbol %s, got %s", expected.Symbol, actual.Symbol)
		t.Fail()
	}
	if expected.Bid != 0 && expected.Bid != actual.Bid {
		t.Errorf("expected bid %f, got %f", expected.Bid, actual.Bid)
		t.Fail()
	}
	if expected.Ask != 0 && expected.Ask != actual.Ask {
		t.Errorf("expected ask %f, got %f", expected.Ask, actual.Ask)
		t.Fail()
	}
	if expected.BidSize != 0 && expected.BidSize != actual.BidSize {
		t.Errorf("expected bid size %f, got %f", expected.BidSize, actual.BidSize)
		t.Fail()
	}
	if expected.Bid != 0 && expected.Bid != actual.Bid {
		t.Errorf("expected bid %f, got %f", expected.Bid, actual.Bid)
		t.Fail()
	}

}

func assert(t *testing.T, expected interface{}, actual interface{}) {
	exp := reflect.ValueOf(expected).Elem()
	act := reflect.ValueOf(actual).Elem()

	if exp.Type() != act.Type() {
		t.Errorf("expected type %s, got %s", exp.Type(), act.Type())
		t.Fail()
	}

	for i := 0; i < exp.NumField(); i++ {
		expValueField := exp.Field(i)
		expTypeField := exp.Type().Field(i)
		//expTag := expTypeField.Tag

		actValueField := act.Field(i)
		actTypeField := act.Type().Field(i)
		//actTag := actTypeField.Tag

		if expTypeField.Name != actTypeField.Name {
			t.Errorf("expected type %s, got %s", expTypeField.Name, actTypeField.Name)
			t.Fail()
		}
		if expValueField.Interface() != actValueField.Interface() {
			t.Errorf("expected %s %s, got %s", expTypeField.Name, expValueField.Interface(), actValueField.Interface())
			t.Fail()
		}
		/*
			if expTag.Get("tag_name") != actTag.Get("tag_name") {
				t.Errorf("expected %s %s, got %s", exp.Type(), expTag.Get("tag_name"), actTag.Get("tag_name"))
			}
		*/
	}
}

func assertSub(t *testing.T, expected *websocket.SubscribeEvent, actual *websocket.SubscribeEvent) {

}

func assertUnsub(t *testing.T, expected *websocket.UnsubscribeEvent, actual *websocket.UnsubscribeEvent) {

}

/*
	Symbol          string
	Bid             float64
	BidPeriod       int64
	BidSize         float64
	Ask             float64
	AskPeriod       int64
	AskSize         float64
	DailyChange     float64
	DailyChangePerc float64
	LastPrice       float64
	Volume          float64
	High            float64
	Low             float64
*/
func TestTicker(t *testing.T) {
	// test setup
	rawInfoEvent := []byte{123, 34, 101, 118, 101, 110, 116, 34, 58, 34, 105, 110, 102, 111, 34, 44, 34, 118, 101, 114, 115, 105, 111, 110, 34, 58, 50, 125}
	rawSubAck := []byte{123, 34, 101, 118, 101, 110, 116, 34, 58, 34, 115, 117, 98, 115, 99, 114, 105, 98, 101, 100, 34, 44, 34, 99, 104, 97, 110, 110, 101, 108, 34, 58, 34, 116, 105, 99, 107, 101, 114, 34, 44, 34, 99, 104, 97, 110, 73, 100, 34, 58, 53, 44, 34, 115, 121, 109, 98, 111, 108, 34, 58, 34, 116, 66, 84, 67, 85, 83, 68, 34, 44, 34, 115, 117, 98, 73, 100, 34, 58, 34, 49, 53, 49, 52, 52, 48, 49, 49, 55, 51, 48, 48, 49, 34, 44, 34, 112, 97, 105, 114, 34, 58, 34, 66, 84, 67, 85, 83, 68, 34, 125}
	rawTick := []byte{91, 53, 44, 91, 49, 52, 57, 53, 55, 44, 54, 56, 46, 49, 55, 51, 50, 56, 55, 57, 54, 44, 49, 52, 57, 53, 56, 44, 53, 53, 46, 50, 57, 53, 56, 56, 49, 51, 50, 44, 45, 54, 53, 57, 44, 45, 48, 46, 48, 52, 50, 50, 44, 49, 52, 57, 55, 49, 44, 53, 51, 55, 50, 51, 46, 48, 56, 56, 49, 51, 57, 57, 53, 44, 49, 54, 52, 57, 52, 44, 49, 52, 52, 53, 52, 93, 93}
	//rawUnsubAck := {91 53 44 91 49 52 57 54 50 44 53 55 46 56 52 57 51 53 48 56 51 44 49 52 57 54 57 44 53 55 46 48 55 53 51 50 55 51 54 44 45 54 54 49 44 45 48 46 48 52 50 51 44 49 52 57 54 57 44 53 51 55 50 50 46 56 51 48 48 55 48 53 54 44 49 54 52 57 52 44 49 52 52 53 52 93 93}
	expectedSub := &websocket.SubscribeEvent{}
	expectedTick := &bitfinex.Ticker{
		Symbol:          "tBTCUSD",
		Bid:             16000.00,
		Ask:             16000.05,
		BidSize:         100,
		AskSize:         105,
		DailyChange:     -160.41,
		DailyChangePerc: -0.1,
		LastPrice:       16000.02,
		Volume:          187325.174673,
		High:            16021.12,
		Low:             15927.85,
	}
	expectedUnsub := &websocket.UnsubscribeEvent{}
	async := newTestAsync()
	fail := make(chan error)
	nonce := &MockNonceGenerator{}

	// create client
	ws := websocket.NewClientWithAsyncNonce(async, nonce)

	// setup listener
	infos := make(chan *websocket.InfoEvent)
	ticks := make(chan *bitfinex.Ticker)
	subs := make(chan *websocket.SubscribeEvent)
	unsubs := make(chan *websocket.UnsubscribeEvent)
	errch := make(chan error)
	go func() {
		for {
			select {
			case msg := <-ws.Listen():
				if msg == nil {
					return
				}
				log.Printf("got msg: %#v", msg)
				// remove threading guarantees when mulitplexing into channels
				switch msg.(type) {
				case error:
					errch <- msg.(error)
				case *bitfinex.Ticker:
					log.Print("sending to ticks")
					ticks <- msg.(*bitfinex.Ticker)
				case *websocket.InfoEvent:
					log.Printf("got info")
					infos <- msg.(*websocket.InfoEvent)
					log.Printf("after info")
				case *websocket.SubscribeEvent:
					subs <- msg.(*websocket.SubscribeEvent)
				case *websocket.UnsubscribeEvent:
					unsubs <- msg.(*websocket.UnsubscribeEvent)
				}
			}
		}
	}()
	ws.SetReadTimeout(time.Second * 2)
	ws.Connect()
	defer ws.Close()
	async.Publish(rawInfoEvent)
	nonce.Next("1514401173001")

	// subscribe
	id, err := ws.SubscribeTicker(context.Background(), "tBTCUSD")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	async.Publish(rawSubAck)
	async.Publish(rawTick)

	// recv & assert
	select {
	case tick := <-ticks:
		assertTicker(t, expectedTick, tick)
	case sub := <-subs:
		assertSub(t, expectedSub, sub)
	case unsub := <-unsubs:
		assertUnsub(t, expectedUnsub, unsub)
	case err := <-fail:
		t.Error(err)
		t.Fail()
	}

	// unsubscribe
	ws.Unsubscribe(context.Background(), id)
	//async.PublishString(rawUnsubAck)
}
