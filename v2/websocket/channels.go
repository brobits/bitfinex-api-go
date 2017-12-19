package websocket

import (
	"github.com/bitfinexcom/bitfinex-api-go/v2/domain"
	"fmt"
	"encoding/json"
)

func (c Client) handleChannel(msg []byte) error {
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
	_, err = c.subscriptions.LookupByChannelID(chanID)
	if err != nil {
		// no subscribed channel for message
		return err
	}
	// TODO use subscription from channel ID lookup below

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
					_, err := c.handlePrivateDataMessage(arr)
					if err != nil {
						return err
					}
					// TODO write private data message strongly typed object to subscription channel
					//sub.Publish(obj)
				}
			}
		}
	case []interface{}:
		// unauthenticated data slice
		// 'data' is data slice
		// 'public' data
		_, err := c.processDataSlice(data)
		if err != nil {
			return err
		}
		//sub.Publish(obj)
		// TODO route data slice object to subscription channel
	}

	return nil
}

func (c Client) handleHeartbeat() {
	// TODO
}

type unsubscribeMsg struct {
	Event  string `json:"event"`
	ChanID int64  `json:"chanId"`
}

// inputs: [ChanID, [Data]], [ChanID, "hb"]
func (c Client) handlePublicDataMessage(raw []interface{}) (interface{}, error) {
	switch len(raw) {
	case 2:
		// [ChanID, [Data]] or [ChanID, "hb"]
		// Data can be either []float64 or [][]float64, where the former should be
		// representing an update and the latter a snapshot.
		// Simple update/snapshot for ticker, books, raw books and candles.
		switch fp := raw[1].(type) {
		case []interface{}:
			return c.processDataSlice(fp)
		case string: // This should be a heartbeat.
			return domain.Heartbeat{}, nil
		}
	case 3:
		// [ChanID, MsgType, [Data]]
		// Data can be either []float64 or [][]float64, where the former should be
		// representing an update and the latter a snapshot.
		if fp, ok := raw[2].([]interface{}); ok {
			return c.processDataSlice(fp)
		}
	}

	return nil, fmt.Errorf("unexpected data message: %#v", raw)
}

func (c Client) processDataSlice(data []interface{}) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("unexpected data slice: %v", data)
	}

	var items [][]float64
	switch data[0].(type) {
	case []interface{}: // [][]float64
		for _, e := range data {
			if s, ok := e.([]interface{}); ok {
				item, err := domain.F64Slice(s)
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			} else {
				return nil, fmt.Errorf("expected slice of float64 slices but got: %v", data)
			}
		}
	case float64: // []float64
		item, err := domain.F64Slice(data)
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
func (c Client) handlePrivateDataMessage(data []interface{}) (ms interface{}, err error) {
	if len(data) < 2 {
		return ms, fmt.Errorf("data message too short: %#v", data)
	}

	term, ok := data[1].(string)
	if !ok {
		return ms, fmt.Errorf("expected data term string in second position but got %#v in %#v", data[1], data)
	}

	if len(data) == 2 || term == "hb" { // Heartbeat
		// TODO: Consider adding a switch to enable/disable passing these along.
		return domain.Heartbeat{}, nil
	}

	list, ok := data[2].([]interface{})
	if !ok {
		return ms, fmt.Errorf("expected data list in third position but got %#v in %#v", data[2], data)
	}

	ms = c.convertRaw(term, list)

	return
}

// convertRaw takes a term and the raw data attached to it to try and convert that
// untyped list into a proper type.
func (c Client) convertRaw(term string, raw []interface{}) interface{} {
	// The things you do to get proper types.
	switch term {
	case "bu":
		o, err := domain.NewBalanceInfoFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.BalanceUpdate(o)
	case "ps":
		o, err := domain.NewPositionSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "pn":
		o, err := domain.NewPositionFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.PositionNew(o)
	case "pu":
		o, err := domain.NewPositionFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.PositionUpdate(o)
	case "pc":
		o, err := domain.NewPositionFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.PositionCancel(o)
	case "ws":
		o, err := domain.NewWalletSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "wu":
		o, err := domain.NewWalletFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.WalletUpdate(o)
	case "os":
		o, err := domain.NewOrderSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "on":
		o, err := domain.NewOrderFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.OrderNew(o)
	case "ou":
		o, err := domain.NewOrderFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.OrderUpdate(o)
	case "oc":
		o, err := domain.NewOrderFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.OrderCancel(o)
	case "hts":
		o, err := domain.NewTradeSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.HistoricalTradeSnapshot(o)
	case "te":
		o, err := domain.NewTradeExecutionFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "tu":
		o, err := domain.NewTradeFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.TradeUpdate(o)
	case "fte":
		o, err := domain.NewFundingTradeFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.FundingTradeExecution(o)
	case "ftu":
		o, err := domain.NewFundingTradeFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.FundingTradeUpdate(o)
	case "hfts":
		o, err := domain.NewFundingTradeSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.HistoricalFundingTradeSnapshot(o)
	case "n":
		o, err := domain.NewNotificationFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "fos":
		o, err := domain.NewFundingOfferSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "fon":
		o, err := domain.NewOfferFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.FundingOfferNew(o)
	case "fou":
		o, err := domain.NewOfferFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.FundingOfferUpdate(o)
	case "foc":
		o, err := domain.NewOfferFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.FundingOfferCancel(o)
	case "fiu":
		o, err := domain.NewFundingInfoFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "fcs":
		o, err := domain.NewFundingCreditSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "fcn":
		o, err := domain.NewCreditFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.FundingCreditNew(o)
	case "fcu":
		o, err := domain.NewCreditFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.FundingCreditUpdate(o)
	case "fcc":
		o, err := domain.NewCreditFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.FundingCreditCancel(o)
	case "fls":
		o, err := domain.NewFundingLoanSnapshotFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	case "fln":
		o, err := domain.NewLoanFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.FundingLoanNew(o)
	case "flu":
		o, err := domain.NewLoanFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.FundingLoanUpdate(o)
	case "flc":
		o, err := domain.NewLoanFromRaw(raw)
		if err != nil {
			return err
		}
		return domain.FundingLoanCancel(o)
		//case "uac":
	case "hb":
		return domain.Heartbeat{}
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
		o, err := domain.NewMarginInfoFromRaw(raw)
		if err != nil {
			return err
		}
		return o
	default:
	}

	return fmt.Errorf("term %q not recognized", term)
}
