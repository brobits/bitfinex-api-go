package websocket

import (
	"sync"
	"errors"
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/utils"
)

type subscriptionRequest struct {
	SubID		string	`json:"subId"`
	Event		string 	`json:"event"`

	// authenticated
	APIKey      string   `json:"apiKey,omitempty"`
	AuthSig     string   `json:"authSig,omitempty"`
	AuthPayload string   `json:"authPayload,omitempty"`
	AuthNonce   string   `json:"authNonce,omitempty"`
	Filter      []string `json:"filter,omitempty"`

	// unauthenticated
	Channel   string `json:"channel,omitempty"`
	Symbol    string `json:"symbol,omitempty"`
	Precision string `json:"prec,omitempty"`
	Frequency string `json:"freq,omitempty"`
	Key       string `json:"key,omitempty"`
	Len       string `json:"len,omitempty"`
	Pair      string `json:"pair,omitempty"`
}

type unsubscribeRequest struct {
	Event  string `json:"event"`
	ChanID int64  `json:"chanId"`
}

type subscription struct {
	ChanID		int64
	pending		bool

	Request 	*subscriptionRequest
	pump 		chan []interface{}
}

func newSubscription(request *subscriptionRequest) *subscription {
	return &subscription{
		Request: request,
		pending: true,
		pump: make(chan []interface{}),
	}
}

func (s subscription) SubID() string {
	return s.Request.SubID
}

// receiver only
func (s subscription) Stream() <-chan []interface{} {
	return s.pump
}

func (s subscription) Publish(msg []interface{}) {
	s.pump <- msg
}

func (s subscription) Pending() bool {
	return s.pending
}

func (s subscription) Close() {
	close(s.pump)
}

type Subscriptions struct {
	lock			sync.Mutex

	subsBySubID		map[string]*subscription // subscription map indexed by subscription ID
	subsByChanID 	map[int64]*subscription // subscription map indexed by channel ID
}

func (s Subscriptions) NextSubID() string {
	return utils.GetNonce()
}

func (s Subscriptions) Add(sub *subscriptionRequest) *subscription {
	s.lock.Lock()
	defer s.lock.Unlock()
	subscription := newSubscription(sub)
	s.subsBySubID[sub.SubID] = subscription
	return subscription
}

func (s Subscriptions) RemoveByChanID(chanID int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	sub, ok := s.subsByChanID[chanID]
	if !ok {
		return errors.New(fmt.Sprintf("could not find channel ID %d", chanID))
	}
	delete(s.subsByChanID, chanID)
	if _, ok = s.subsBySubID[sub.SubID()]; ok {
		delete(s.subsBySubID, sub.SubID())
	}
	sub.Close()
	return nil
}

func (s Subscriptions) RemoveBySubID(subID string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	sub, ok := s.subsBySubID[subID]
	if !ok {
		return errors.New(fmt.Sprintf("could not find subscription ID %s", subID))
	}
	// exists, remove both indices
	delete(s.subsBySubID, subID)
	if _, ok = s.subsByChanID[sub.ChanID]; ok {
		delete(s.subsByChanID, sub.ChanID)
	}
	sub.Close()
	return nil
}

func (s Subscriptions) Activate(subID string, chanID int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if sub, ok := s.subsBySubID[subID]; ok {
		sub.pending = false
		s.subsByChanID[chanID] = sub
		return nil
	}
	return errors.New(fmt.Sprintf("could not find subscription ID %s", subID))
}

func (s Subscriptions) Route(chanID int64, msg []interface{}) error {
	sub, err := s.LookupByChannelID(chanID)
	if err != nil {
		return err
	}
	sub.Publish(msg)
	return nil
}

func (s Subscriptions) LookupByChannelID(chanID int64) (*subscription, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if sub, ok := s.subsByChanID[chanID]; ok {
		return sub, nil
	}
	return nil, errors.New(fmt.Sprintf("could not find subscription for channel ID %d", chanID))
}

func (s Subscriptions) LookupBySubscriptionID(subID string) (*subscription, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if sub, ok := s.subsBySubID[subID]; ok {
		return sub, nil
	}
	return nil, errors.New(fmt.Sprintf("could not find subscription ID %s", subID))
}