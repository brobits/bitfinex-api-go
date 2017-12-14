package domain

import "fmt"

type Ticker struct {
	Symbol          string
	Bid             float64
	BidPeriod       int64
	BidSize         float64
	Ask             float64
	AskPeriod       int64
	AskSize         float64
	DailyChange     float64
	DailyChangePerc float64
	LastPrice       float64
	Volume          float64
	High            float64
	Low             float64
}

type TickerUpdate Ticker
type TickerSnapshot []Ticker

//type Trade struct {
//ID     int64
//MTS    int64
//Amount float64
//Price  float64
//Rate   float64
//Period int64
//}

func NewTickerFromRaw(raw []interface{}) (t Ticker, err error) {
	if len(raw) < 26 {
		return t, fmt.Errorf("data slice too short for ticker: %#v", raw)
	}

	t = Ticker{
		/*
		ID:            int64(f64ValOrZero(raw[0])),
		GID:           int64(f64ValOrZero(raw[1])),
		CID:           int64(f64ValOrZero(raw[2])),
		Symbol:        sValOrEmpty(raw[3]),
		MTSCreated:    int64(f64ValOrZero(raw[4])),
		MTSUpdated:    int64(f64ValOrZero(raw[5])),
		Amount:        f64ValOrZero(raw[6]),
		AmountOrig:    f64ValOrZero(raw[7]),
		Type:          sValOrEmpty(raw[8]),
		TypePrev:      sValOrEmpty(raw[9]),
		Flags:         i64ValOrZero(raw[12]),
		Status:        OrderStatus(sValOrEmpty(raw[13])),
		Price:         f64ValOrZero(raw[16]),
		PriceAvg:      f64ValOrZero(raw[17]),
		PriceTrailing: f64ValOrZero(raw[18]),
		PriceAuxLimit: f64ValOrZero(raw[19]),
		Notify:        bValOrFalse(raw[23]),
		Hidden:        bValOrFalse(raw[24]),
		PlacedID:      i64ValOrZero(raw[25]),*/
	}

	return
}