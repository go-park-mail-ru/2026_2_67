package usecase

import (
	"context"
	"vibe_market/backend/internal/pickup_points/domain"
)

type PickupPointUsecase interface {
	PickupPoints(ctx context.Context) (PickupPointsOutDTO, error)
}

type PickupPointUsecaseImpl struct {
	pickupPointRepo domain.PickupPointRepo
}

func NewPickupPointUsecase(pickupPointRepo domain.PickupPointRepo) *PickupPointUsecaseImpl {
	return &PickupPointUsecaseImpl{pickupPointRepo}
}
