package rest

import (
	"github.com/bitfinexcom/bitfinex-api-go/v2/domain"
)

// PositionService manages the Position endpoint.
type PositionService struct {
	Synchronous
}

// All returns all positions for the authenticated account.
func (s *PositionService) All() (domain.PositionSnapshot, error) {
	raw, err := s.Request(NewRequest("positions"))

	if err != nil {
		return nil, err
	}

	os, err := domain.NewPositionSnapshotFromRaw(raw)
	if err != nil {
		return nil, err
	}

	return os, nil
}
