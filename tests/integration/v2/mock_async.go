package tests

import (
	"context"
	"errors"
	"log"
)

type TestAsync struct {
	bridge    chan []byte
	connected bool
}

func (t *TestAsync) Connect() error {
	t.connected = true
	return nil
}

func (t *TestAsync) Send(ctx context.Context, msg interface{}) error {
	if !t.connected {
		return errors.New("must connect before sending")
	}
	return nil
}

func (t *TestAsync) Listen() <-chan []byte {
	return t.bridge
}

func (t *TestAsync) PublishString(raw string) {
	b := []byte(raw)
	log.Printf("mock publish:\n%s", b)
	t.bridge <- b
}

func (t *TestAsync) Publish(raw []byte) {
	t.bridge <- raw
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
	}
}
