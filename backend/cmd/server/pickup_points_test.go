package main

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestPickupPointsRegistration(t *testing.T) {
	for _, tt := range []struct {
		method string
		path   string
		status int
		calls  int
	}{
		{http.MethodGet, "/api/v1/pickup_points", http.StatusOK, 1},
		{http.MethodGet, "/api/v1/pickup_points/", http.StatusOK, 1},
		{http.MethodPost, "/api/v1/pickup_points", http.StatusMethodNotAllowed, 0},
		{http.MethodPost, "/api/v1/pickup_points/", http.StatusMethodNotAllowed, 0},
		{http.MethodPut, "/api/v1/pickup_points", http.StatusMethodNotAllowed, 0},
		{http.MethodPut, "/api/v1/pickup_points/", http.StatusMethodNotAllowed, 0},
		{http.MethodDelete, "/api/v1/pickup_points", http.StatusMethodNotAllowed, 0},
		{http.MethodDelete, "/api/v1/pickup_points/", http.StatusMethodNotAllowed, 0},
		{http.MethodGet, "/api/v1/pickup_points/unknown", http.StatusNotFound, 0},
		{http.MethodGet, "/api/v1/pickup_points/unknown/", http.StatusNotFound, 0},
		{http.MethodGet, "/api/v1/pickup_points_extra", http.StatusNotFound, 0},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "request value")
			calls := 0
			pickupPointRepo := pickupPointRepoStub{list: func(got context.Context) ([]domain.PickupPoint, error) {
				calls++
				if got != ctx || got.Value(contextKey{}) != "request value" {
					t.Error("registration did not preserve request context")
				}
				return []domain.PickupPoint{
					{ID: 17, Longitude: 37.6173, Latitude: 55.7558},
					{ID: 23, Longitude: -73.5, Latitude: 0},
				}, nil
			}}
			logger := zap.NewNop()
			mainMux := http.NewServeMux()

			// Keep this registration block identical to cmd/server/main.go.
			pickupPointsMux := pickup_points.NewServeMux(pickupPointRepo, logger)
			pickupPointsHandler := http.StripPrefix("/api/v1/pickup_points", pickupPointsMux)
			mainMux.HandleFunc("/api/v1/pickup_points", func(w http.ResponseWriter, r *http.Request) {
				r = r.Clone(r.Context())
				r.URL.Path = "/api/v1/pickup_points/"
				r.URL.RawPath = ""
				pickupPointsHandler.ServeHTTP(w, r)
			})
			mainMux.Handle("/api/v1/pickup_points/{$}", pickupPointsHandler)

			request := httptest.NewRequest(tt.method, tt.path, nil).WithContext(ctx)
			originalURL := *request.URL
			recorder := httptest.NewRecorder()
			mainMux.ServeHTTP(recorder, request)
			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d; body: %s", recorder.Code, tt.status, recorder.Body.String())
			}
			if location := recorder.Header().Get("Location"); location != "" {
				t.Errorf("unexpected redirect to %q", location)
			}
			if calls != tt.calls {
				t.Errorf("repo calls = %d, want %d", calls, tt.calls)
			}
			if *request.URL != originalURL {
				t.Error("registration mutated the original request URL")
			}
			if tt.status == http.StatusOK {
				if got := recorder.Header().Get("Content-Type"); got != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", got)
				}
				want := `{"pickupPoints":[{"id":17,"longitude":37.6173,"latitude":55.7558},{"id":23,"longitude":-73.5,"latitude":0}]}` + "\n"
				if got := recorder.Body.String(); got != want {
					t.Errorf("body = %q, want %q", got, want)
				}
			}
		})
	}
}
