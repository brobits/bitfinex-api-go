package websocket

import (
	"sync"
	"errors"
	"fmt"
)

type Subscription struct {
	ChanID		int64
	pending		bool

	Request 	*PublicSubscriptionRequest
	pump 		chan []interface{}
}

func newSubscription(request *PublicSubscriptionRequest) *Subscription {
	return &Subscription{
		Request: request,
		pending: true,
		pump: make(chan []interface{}),
	}
}

func (s Subscription) SubID() string {
	return s.Request.SubID
}

// receiver only
func (s Subscription) Stream() <-chan []interface{} {
	return s.pump
}

func (s Subscription) Publish(msg []interface{}) {
	s.pump <- msg
}

func (s Subscription) Pending() bool {
	return s.pending
}

func (s Subscription) Close() {
	close(s.pump)
}

type Subscriptions struct {
	lock			sync.Mutex
	subs			int64

	subsBySubID		map[string]*Subscription // subscription map indexed by subscription ID
	subsByChanID 	map[int64]*Subscription // subscription map indexed by channel ID
}

func (s Subscriptions) NextSubID() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.subs++
	return fmt.Sprintf("%d", s.subs)
}

func (s Subscriptions) Add(sub *PublicSubscriptionRequest) {
	s.lock.Lock()
	defer s.lock.Unlock()
	subscription := newSubscription(sub)
	s.subsBySubID[sub.SubID] = subscription
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

func (s Subscriptions) LookupByChannelID(chanID int64) (*Subscription, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if sub, ok := s.subsByChanID[chanID]; ok {
		return sub, nil
	}
	return nil, errors.New(fmt.Sprintf("could not find subscription for channel ID %d", chanID))
}

func (s Subscriptions) LookupBySubscriptionID(subID string) (*Subscription, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if sub, ok := s.subsBySubID[subID]; ok {
		return sub, nil
	}
	return nil, errors.New(fmt.Sprintf("could not find subscription ID %s", subID))
}