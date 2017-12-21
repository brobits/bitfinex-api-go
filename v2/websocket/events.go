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

type RawEvent struct {
	Data interface{}
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

// onEvent handles all the event messages and connects SubID and ChannelID.
func (c *Client) handleEvent(msg []byte) error {
	event := &eventType{}
	err := json.Unmarshal(msg, event)
	if err != nil {
		return err
	}

	//var e interface{}
	switch event.Event {
	case "info":
		i := InfoEvent{}
		// TODO e->i
		c.listener <- i
	case "auth":
		a := AuthEvent{}
		err = json.Unmarshal(msg, &a)
		if err != nil {
			return err
		}
		c.subscriptions.activate(a.SubID, a.ChanID)
		c.Authentication = SuccessfulAuthentication
		c.listener <- a
		return nil
	case "subscribed":
		s := SubscribeEvent{}
		err = json.Unmarshal(msg, &s)
		if err != nil {
			return err
		}
		c.subscriptions.activate(s.SubID, s.ChanID)
		c.listener <- s
		return nil
	case "unsubscribed":
		s := UnsubscribeEvent{}
		err = json.Unmarshal(msg, &s)
		if err != nil {
			return err
		}
		c.subscriptions.removeByChanID(s.ChanID)
		c.listener <- s
	case "error":
		er := ErrorEvent{}
		// TODO e->er
		c.listener <- er
	case "conf":
		ec := ConfEvent{}
		// TODO e->ec
		c.listener <- ec
	default:
		return fmt.Errorf("unknown event: %s", msg) // TODO: or just log?
	}

	//err = json.Unmarshal(msg, &e)
	//TODO raw message isn't ever published

	return err
}
