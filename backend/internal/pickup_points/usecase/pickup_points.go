package usecase

import (
	"context"

	"vibe_market/backend/internal/pickup_points/domain"
)

type PickupPointsOutDTO struct {
	PickupPoints []domain.PickupPoint
}

func (u *PickupPointUsecaseImpl) PickupPoints(ctx context.Context) (PickupPointsOutDTO, error) {
	points, err := u.pickupPointRepo.List(ctx)
	if err != nil {
		return PickupPointsOutDTO{}, err
	}
	return PickupPointsOutDTO{PickupPoints: points}, nil
}
