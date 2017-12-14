package websocket

import (
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/v2/domain"
)

func (b *bfxWebsocket) handlePrivateDataMessage(data []interface{}) (ms interface{}, err error) {
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

	ms = b.convertRaw(term, list)

	return
}

// convertRaw takes a term and the raw data attached to it to try and convert that
// untyped list into a proper type.
func (b *bfxWebsocket) convertRaw(term string, raw []interface{}) interface{} {
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
