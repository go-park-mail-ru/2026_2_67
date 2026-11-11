package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"vibe_market/backend/internal/pickup_points/domain"
	"vibe_market/backend/internal/pickup_points/usecase"
)

type pickupPointRepoStub struct {
	list func(context.Context) ([]domain.PickupPoint, error)
}

func (repo pickupPointRepoStub) List(ctx context.Context) ([]domain.PickupPoint, error) {
	return repo.list(ctx)
}

func TestPickupPoints(t *testing.T) {
	repoErr := errors.New("repository failed")
	tests := []struct {
		name   string
		points []domain.PickupPoint
		err    error
	}{
		{name: "success", points: []domain.PickupPoint{{ID: 42, Longitude: 37.6, Latitude: 55.7}}},
		{name: "empty", points: []domain.PickupPoint{}},
		{name: "nil slice"},
		{name: "error", err: repoErr},
		{name: "error with partial result", points: []domain.PickupPoint{{ID: 42}}, err: repoErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			repo := pickupPointRepoStub{list: func(gotCtx context.Context) ([]domain.PickupPoint, error) {
				calls++
				if gotCtx != ctx {
					t.Error("context was not passed unchanged")
				}
				return tt.points, tt.err
			}}
			got, err := usecase.PickupPoints(ctx, repo)
			if err != tt.err {
				t.Fatalf("error = %v, want %v", err, tt.err)
			}
			want := usecase.PickupPointsOutDTO{}
			if tt.err == nil {
				want.PickupPoints = tt.points
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("PickupPoints() = %#v, want %#v", got, want)
			}
			if calls != 1 {
				t.Errorf("List called %d times, want 1", calls)
			}
		})
	}
}
