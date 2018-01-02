package tests

import (
	"context"
	"errors"
	"log"
)

type TestAsync struct {
	bridge    chan []byte
	connected bool
	Sent      []interface{}
}

func (t *TestAsync) Connect() error {
	t.connected = true
	return nil
}

func (t *TestAsync) Send(ctx context.Context, msg interface{}) error {
	if !t.connected {
		return errors.New("must connect before sending")
	}
	t.Sent = append(t.Sent, msg)
	return nil
}

func (t *TestAsync) DumpSentMessages() {
	for i, msg := range t.Sent {
		log.Printf("%2d: %#v", i, msg)
	}
}

func (t *TestAsync) Listen() <-chan []byte {
	return t.bridge
}

func (t *TestAsync) Publish(raw string) {
	t.bridge <- []byte(raw)
}

func (t *TestAsync) Close() {
	close(t.bridge)
}

func (t *TestAsync) Done() <-chan error {
	ch := make(chan error)
	return ch
}

func newTestAsync() *TestAsync {
	return &TestAsync{
		bridge:    make(chan []byte),
		connected: false,
		Sent:      make([]interface{}, 0),
	}
}
