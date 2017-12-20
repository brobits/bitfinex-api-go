package websocket

import (
	"encoding/json"
	"fmt"
)

type eventType struct {
	Event string `json:"event"`
}

type InfoEvent struct {
	Version int `json:"version"`
}

type AuthEvent struct {
	Event   string       `json:"event"`
	Status  string       `json:"status"`
	ChanID  int64        `json:"chanId,omitempty"`
	UserID  int64        `json:"userId,omitempty"`
	SubID   string       `json:"subId"`
	AuthID  string       `json:"auth_id,omitempty"`
	Message string       `json:"msg,omitempty"`
	Caps    Capabilities `json:"caps"`
}

type Capability struct {
	Read  int `json:"read"`
	Write int `json:"write"`
}

type Capabilities struct {
	Orders    Capability `json:"orders"`
	Account   Capability `json:"account"`
	Funding   Capability `json:"funding"`
	History   Capability `json:"history"`
	Wallets   Capability `json:"wallets"`
	Withdraw  Capability `json:"withdraw"`
	Positions Capability `json:"positions"`
}

type ErrorEvent struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

type UnsubscribeEvent struct {
	Status string `json:"status"`
	ChanID int64  `json:"chanId"`
}

type SubscribeEvent struct {
	SubID     string `json:"subId"`
	Channel   string `json:"channel"`
	ChanID    int64  `json:"chanId"`
	Symbol    string `json:"symbol"`
	Precision string `json:"prec,omitempty"`
	Frequency string `json:"freq,omitempty"`
	Key       string `json:"key,omitempty"`
	Len       string `json:"len,omitempty"`
	Pair      string `json:"pair"`
}

type ConfEvent struct {
	Flags int `json:"flags"`
}

type EventListener interface {
	onInfo(e InfoEvent)
	onAuth(e AuthEvent)
	onSubscribe(e SubscribeEvent)
	onUnsubscribe(e UnsubscribeEvent)
	onError(e ErrorEvent)
	onConf(e ConfEvent)
	onRawEvent(e interface{})
}

type DefaultEventListener struct {
	EventListener
}

func (d DefaultEventListener) onInfo(e InfoEvent) {
	// no-op
}

func (d DefaultEventListener) onAuth(e AuthEvent) {
	// no-op
}

func (d DefaultEventListener) onSubscribe(e SubscribeEvent) {
	// no-op
}

func (d DefaultEventListener) onUnsubscribe(e UnsubscribeEvent) {
	// no-op
}

func (d DefaultEventListener) onError(e ErrorEvent) {
	// no-op
}

func (d DefaultEventListener) onConf(e ConfEvent) {
	// no-op
}

func (d DefaultEventListener) onRawEvent(e interface{}) {
	// no-op
}


func (c Client) RegisterEventListener(listener EventListener) {
	c.eventListener = listener
}

// onEvent handles all the event messages and connects SubID and ChannelID.
func (c Client) handleEvent(msg []byte) error {
	event := &eventType{}
	err := json.Unmarshal(msg, event)
	if err != nil {
		return err
	}

	var e interface{}
	switch event.Event {
	case "info":
		i := InfoEvent{}
		// TODO e->i
		if c.eventListener != nil {
			c.eventListener.onInfo(i)
		}
	case "auth":
		a := AuthEvent{}
		err = json.Unmarshal(msg, &a)
		if err != nil {
			return err
		}
		c.subscriptions.Activate(a.SubID, a.ChanID)
		c.Authentication = SuccessfulAuthentication
		if c.eventListener != nil {
			c.eventListener.onAuth(a)
		}
		return nil
	case "subscribed":
		s := SubscribeEvent{}
		err = json.Unmarshal(msg, &s)
		if err != nil {
			return err
		}
		c.subscriptions.Activate(s.SubID, s.ChanID)
		if c.eventListener != nil {
			c.eventListener.onSubscribe(s)
		}
		return nil
	case "unsubscribed":
		s := UnsubscribeEvent{}
		err = json.Unmarshal(msg, &s)
		if err != nil {
			return err
		}
		c.subscriptions.RemoveByChanID(s.ChanID)
		if c.eventListener != nil {
			c.eventListener.onUnsubscribe(s)
		}
	case "error":
		er := ErrorEvent{}
		// TODO e->er
		if c.eventListener != nil {
			c.eventListener.onError(er)
		}
	case "conf":
		ec := ConfEvent{}
		// TODO e->ec
		if c.eventListener != nil {
			c.eventListener.onConf(ec)
		}
	default:
		return fmt.Errorf("unknown event: %s", msg) // TODO: or just log?
	}

	err = json.Unmarshal(msg, &e)
	if err != nil && c.eventListener != nil {
		c.eventListener.onRawEvent(e)
	}

	return err
}
