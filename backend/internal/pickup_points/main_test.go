package pickup_points_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"vibe_market/backend/internal/pickup_points"
	"vibe_market/backend/internal/pickup_points/domain"

	"go.uber.org/zap"
)

type pickupPointRepoStub struct {
	list func(context.Context) ([]domain.PickupPoint, error)
}

var _ domain.PickupPointRepo = pickupPointRepoStub{}

func (s pickupPointRepoStub) List(ctx context.Context) ([]domain.PickupPoint, error) {
	return s.list(ctx)
}

func TestNewServeMuxGet(t *testing.T) {
	tests := []struct {
		name   string
		points []domain.PickupPoint
		want   string
	}{
		{
			name: "points with numeric coordinates",
			points: []domain.PickupPoint{
				{ID: 17, Longitude: 37.6173, Latitude: 55.7558},
				{ID: 23, Longitude: -73.5, Latitude: 0},
			},
			want: `{"pickupPoints":[{"id":17,"longitude":37.6173,"latitude":55.7558},{"id":23,"longitude":-73.5,"latitude":0}]}`,
		},
		{name: "nil slice", want: `{"pickupPoints":[]}`},
		{name: "empty slice", points: []domain.PickupPoint{}, want: `{"pickupPoints":[]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			repo := pickupPointRepoStub{list: func(context.Context) ([]domain.PickupPoint, error) {
				calls++
				return tt.points, nil
			}}
			recorder := httptest.NewRecorder()
			pickup_points.NewServeMux(repo, zap.NewNop()).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			var got, want any
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatalf("decode expected response: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("response = %s, want %s", recorder.Body.String(), tt.want)
			}
			if calls != 1 {
				t.Errorf("repo calls = %d, want 1", calls)
			}
		})
	}
}

func TestNewServeMuxRepoError(t *testing.T) {
	calls := 0
	repo := pickupPointRepoStub{list: func(context.Context) ([]domain.PickupPoint, error) {
		calls++
		return nil, errors.New("repository unavailable")
	}}
	recorder := httptest.NewRecorder()
	pickup_points.NewServeMux(repo, zap.NewNop()).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if calls != 1 {
		t.Errorf("repo calls = %d, want 1", calls)
	}
}

func TestNewServeMuxRejectsRequestsWithoutCallingRepo(t *testing.T) {
	for _, tt := range []struct {
		method string
		path   string
		status int
	}{
		{http.MethodPost, "/", http.StatusMethodNotAllowed},
		{http.MethodPut, "/", http.StatusMethodNotAllowed},
		{http.MethodDelete, "/", http.StatusMethodNotAllowed},
		{http.MethodGet, "/unknown", http.StatusNotFound},
		{http.MethodGet, "/unknown/", http.StatusNotFound},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			repo := pickupPointRepoStub{list: func(context.Context) ([]domain.PickupPoint, error) {
				t.Error("repo must not be called")
				return nil, nil
			}}
			recorder := httptest.NewRecorder()
			pickup_points.NewServeMux(repo, zap.NewNop()).ServeHTTP(recorder, httptest.NewRequest(tt.method, tt.path, nil))
			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d", recorder.Code, tt.status)
			}
		})
	}
}

func TestNewServeMuxPassesRequestContext(t *testing.T) {
	type contextKey struct{}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "request value"))
	defer cancel()
	cancel()
	calls := 0
	repo := pickupPointRepoStub{list: func(got context.Context) ([]domain.PickupPoint, error) {
		calls++
		if got != ctx {
			t.Error("repo did not receive the original request context")
		}
		if got.Value(contextKey{}) != "request value" || got.Err() != context.Canceled {
			t.Error("request context value or cancellation was lost")
		}
		return nil, nil
	}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	pickup_points.NewServeMux(repo, zap.NewNop()).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if calls != 1 {
		t.Errorf("repo calls = %d, want 1", calls)
	}
}
