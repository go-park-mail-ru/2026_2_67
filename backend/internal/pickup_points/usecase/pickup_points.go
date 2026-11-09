package usecase

import (
	"context"

	"vibe_market/backend/internal/pickup_points/domain"
)

type PickupPointsOutDTO struct {
	PickupPoints []domain.PickupPoint
}

func PickupPoints(ctx context.Context, repo domain.PickupPointRepo) (PickupPointsOutDTO, error) {
	points, err := repo.List(ctx)
	if err != nil {
		return PickupPointsOutDTO{}, err
	}
	return PickupPointsOutDTO{PickupPoints: points}, nil
}
