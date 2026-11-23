package domain

import (
	"context"
	"time"
)

type PickupPoint struct {
	ID        int64
	Longitude float64
	Latitude  float64
	StartTime time.Time
	EndTime   time.Time
}

type PickupPointRepo interface {
	List(ctx context.Context) ([]PickupPoint, error)
}
