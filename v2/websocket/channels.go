package websocket

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bitfinexcom/bitfinex-api-go/v2"
)

func (c *Client) handleChannel(msg []byte) error {
	var raw []interface{}
	err := json.Unmarshal(msg, &raw)
	if err != nil {
		return err
	} else if len(raw) < 2 {
		return nil
	}

	chID, ok := raw[0].(float64)
	if !ok {
		return fmt.Errorf("expected message to start with a channel id but got %#v instead", raw[0])
	}

	chanID := int64(chID)
	sub, err := c.subscriptions.lookupByChannelID(chanID)
	if err != nil {
		// no subscribed channel for message
		return err
	}

	// public msg: [ChanID, [Data]]
	// hb (both): [ChanID, "hb"]
	// private msg: [ChanID, "type", [Data]]
	switch data := raw[1].(type) {
	case string:
		// authenticated data slice, or a heartbeat
		if raw[1].(string) == "hb" {
			c.handleHeartbeat()
		} else {
			// authenticated data slice
			// raw[2] is data slice
			// 'private' data
			if len(raw) > 2 {
				if arr, ok := raw[2].([]interface{}); ok {
					obj, err := c.handlePrivateDataMessage(raw[1].(string), arr)
					if err != nil {
						return err
					}
					// private data is returned as strongly typed data, publish directly
					c.listener <- obj
				}
			}
		}
	case []interface{}:
		// unauthenticated data slice
		// 'data' is data slice
		// 'public' data
		// returns interface{} (which is really [][]float64)
		obj, err := c.processDataSlice(data)
		if err != nil {
			return err
		}
		// public data is returned as raw interface arrays, use a factory to convert to raw type & publish
		if factory, ok := c.factories[sub.Request.Channel]; ok {
			flt := obj.([][]float64)
			if len(flt) >= 1 {
				arr := make([]interface{}, len(flt[0]))
				for i, ft := range flt[0] {
					arr[i] = ft
				}
				msg, err := factory(chanID, arr)
				if err != nil {
					// factory error
					return err
				}
				c.listener <- msg
			} else {
				log.Printf("data too small to process: %#v", obj)
			}
		} else {
			// factory lookup error
			return fmt.Errorf("could not find public factory for %s channel", sub.Request.Channel)
		}
	}

	return nil
}

func (c *Client) handleHeartbeat() {
	// TODO internal heartbeat timeout thread?
}

type unsubscribeMsg struct {
	Event  string `json:"event"`
	ChanID int64  `json:"chanId"`
}

func (c *Client) processDataSlice(data []interface{}) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("unexpected data slice: %v", data)
	}

	var items [][]float64
	switch data[0].(type) {
	case []interface{}: // [][]float64
		for _, e := range data {
			if s, ok := e.([]interface{}); ok {
				item, err := bitfinex.F64Slice(s)
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			} else {
				return nil, fmt.Errorf("expected slice of float64 slices but got: %v", data)
			}
		}
	case float64: // []float64
		item, err := bitfinex.F64Slice(data)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	default:
		return nil, fmt.Errorf("unexpected data slice: %v", data)
	}

	return items, nil
}

// public msg: [ChanID, [Data]]
// hb (both): [ChanID, "hb"]
// private msg: [ChanID, "type", [Data]]
func (c *Client) handlePrivateDataMessage(term string, data []interface{}) (ms interface{}, err error) {
	if len(data) < 2 {
		return ms, fmt.Errorf("data message too short: %#v", data)
	}
	/*
		term, ok := data[1].(string)
		// TODO trades violating the following check
		if !ok {
			return ms, fmt.Errorf("expected data term string in second position but got %#v in %#v", data[1], data)
		}
	*/
	if len(data) == 1 || term == "hb" { // Heartbeat
		// TODO: Consider adding a switch to enable/disable passing these along.
		return bitfinex.Heartbeat{}, nil
	}
	/*
		list, ok := data[2].([]interface{})
		if !ok {
			return ms, fmt.Errorf("expected data list in third position but got %#v in %#v", data[2], data)
		}
	*/
	ms = c.convertRaw(term, data)

	return
}

// convertRaw takes a term and the raw data attached to it to try and convert that
// untyped list into a proper type.
func (c *Client) convertRaw(term string, raw []interface{}) interface{} {
	// The things you do to get proper types.
	switch term {
	case "bu":
		o, err := bitfinex.NewBalanceInfoFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.BalanceUpdate(o)
	case "ps":
		o, err := bitfinex.NewPositionSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "pn":
		o, err := bitfinex.NewPositionFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.PositionNew(o)
	case "pu":
		o, err := bitfinex.NewPositionFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.PositionUpdate(o)
	case "pc":
		o, err := bitfinex.NewPositionFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.PositionCancel(o)
	case "ws":
		o, err := bitfinex.NewWalletSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "wu":
		o, err := bitfinex.NewWalletFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.WalletUpdate(o)
	case "os":
		o, err := bitfinex.NewOrderSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "on":
		o, err := bitfinex.NewOrderFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.OrderNew(o)
	case "ou":
		o, err := bitfinex.NewOrderFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.OrderUpdate(o)
	case "oc":
		o, err := bitfinex.NewOrderFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.OrderCancel(o)
	case "hts":
		o, err := bitfinex.NewTradeSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.HistoricalTradeSnapshot(o)
	case "te":
		o, err := bitfinex.NewTradeExecutionFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "tu":
		o, err := bitfinex.NewTradeFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.TradeUpdate(o)
	case "fte":
		o, err := bitfinex.NewFundingTradeFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.FundingTradeExecution(o)
	case "ftu":
		o, err := bitfinex.NewFundingTradeFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.FundingTradeUpdate(o)
	case "hfts":
		o, err := bitfinex.NewFundingTradeSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.HistoricalFundingTradeSnapshot(o)
	case "n":
		o, err := bitfinex.NewNotificationFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "fos":
		o, err := bitfinex.NewFundingOfferSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "fon":
		o, err := bitfinex.NewOfferFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.FundingOfferNew(o)
	case "fou":
		o, err := bitfinex.NewOfferFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.FundingOfferUpdate(o)
	case "foc":
		o, err := bitfinex.NewOfferFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.FundingOfferCancel(o)
	case "fiu":
		o, err := bitfinex.NewFundingInfoFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "fcs":
		o, err := bitfinex.NewFundingCreditSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "fcn":
		o, err := bitfinex.NewCreditFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.FundingCreditNew(o)
	case "fcu":
		o, err := bitfinex.NewCreditFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.FundingCreditUpdate(o)
	case "fcc":
		o, err := bitfinex.NewCreditFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.FundingCreditCancel(o)
	case "fls":
		o, err := bitfinex.NewFundingLoanSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "fln":
		o, err := bitfinex.NewLoanFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.FundingLoanNew(o)
	case "flu":
		o, err := bitfinex.NewLoanFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.FundingLoanUpdate(o)
	case "flc":
		o, err := bitfinex.NewLoanFromRaw(raw)
		if err != nil {
			return err
		}
		return bitfinex.FundingLoanCancel(o)
		//case "uac":
	case "hb":
		return bitfinex.Heartbeat{}
	case "ats":
		// TODO: Is not in documentation, so figure out what it is.
		return nil
	case "oc-req":
		// TODO
		return nil
	case "on-req":
		// TODO
		return nil
	case "mis": // Should not be sent anymore as of 2017-04-01
		return nil
	case "miu":
		o, err := bitfinex.NewMarginInfoFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	default:
	}

	return fmt.Errorf("term %q not recognized", term)
}
