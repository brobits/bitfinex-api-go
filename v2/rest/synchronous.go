package rest

import "github.com/bitfinexcom/bitfinex-api-go/v2/domain"

type SynchronousRequest struct {
	RefURL string
	Data map[string]interface{}
}

type Synchronous interface {
	Request(request interface{}) []interface{}
}

type OrderProvider interface {
	All() (domain.OrderSnapshot, error)
	History() (domain.OrderSnapshot, error)
	Status(orderID int64) (domain.Order, error)

}

