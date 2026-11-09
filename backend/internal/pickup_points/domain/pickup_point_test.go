package domain_test

import (
	"context"
	"testing"

	"vibe_market/backend/internal/pickup_points/domain"
)

type pickupPointRepoFunc func(context.Context) ([]domain.PickupPoint, error)

func (f pickupPointRepoFunc) List(ctx context.Context) ([]domain.PickupPoint, error) {
	return f(ctx)
}

func TestPickupPointRepoContract(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	want := domain.PickupPoint{ID: 1 << 40, Longitude: -73.9857, Latitude: 40.7484}
	var repo domain.PickupPointRepo = pickupPointRepoFunc(func(gotCtx context.Context) ([]domain.PickupPoint, error) {
		if gotCtx != ctx {
			t.Error("context was not passed unchanged")
		}
		return []domain.PickupPoint{want}, nil
	})
	points, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 || points[0] != want {
		t.Fatalf("List() = %#v, want [%#v]", points, want)
	}
}
