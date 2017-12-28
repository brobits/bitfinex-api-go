package tests

import (
	"context"
	"errors"
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
	}
}
