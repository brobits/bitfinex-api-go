package bitfinex

import (
	"net/url"
)

type TickerService struct {
	client *Client
}

// Status indicates whether the platform is currently operative or not.
func (p *TickerService) Get(symbol string) ([]interface{}, error) {
	params := url.Values{}
	params.Add("symbols", "tBTCUSD")
	req, err := p.client.newRequest("GET", "tickers", params, nil)
	if err != nil {
		return nil, err
	}

	var s []interface{}
	_, err = p.client.do(req, &s)
	if err != nil {
		return nil, err
	}

	return s, nil
}
