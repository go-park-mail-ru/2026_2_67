package domain

import "context"

type PickupPoint struct {
	ID        int64
	Longitude float64
	Latitude  float64
}

type PickupPointRepo interface {
	List(ctx context.Context) ([]PickupPoint, error)
}
