package features

import (
	. "github.com/gucumber/gucumber"
	"log"
	"encoding/json"
	"io/ioutil"
	"bytes"
	"net/http"
	"github.com/bitfinexcom/bitfinex-api-go/v2"
)

func init() {

	var httpDo func(client *http.Client, req *http.Request) (*http.Response, error)
	var expected []interface{}
	var actual []interface{}

	Given(`^I preload the URL "(.+)" with the following JSON:$`, func(url string, data string) {
		err := json.Unmarshal([]byte(data), &expected)
		if err != nil {
			log.Panic(err)
		}
		httpDo = func(client *http.Client, req *http.Request) (*http.Response, error) {
			resp := http.Response{
				Body:       ioutil.NopCloser(bytes.NewBufferString(data)),
				StatusCode: 200,
			}
			return &resp, nil
		}
	})

	When(`^I visit the URL "(.+?)"$`, func(url string) {
		ticker, err := bitfinex.NewClientWithHTTPDo(httpDo).Ticker.Get("tBTCUSD")
		if err != nil {
			log.Panic(err)
		}
		actual = ticker
	})

	Then(`^I should see the following JSON:$`, func(data string) {
		err := json.Unmarshal([]byte(data), &actual)
		if err != nil {
			log.Panic(err)
		}
		for i, e := range expected {
			a := actual[i]
			if e != a {
				log.Panicf("expected %x, got %x", e, a)
			}
		}
	})
}