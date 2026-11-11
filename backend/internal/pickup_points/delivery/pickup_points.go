package delivery

import (
	"encoding/json"
	"net/http"
	"vibe_market/backend/internal/pickup_points/usecase"
)

type pickupPointJSON struct {
	ID        int64   `json:"id"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type pickupPointsResponse struct {
	PickupPoints []pickupPointJSON `json:"pickupPoints"`
}

func (h *PickupPointsHandler) PickupPoints(w http.ResponseWriter, r *http.Request) {
	out, err := usecase.PickupPoints(r.Context(), h.pickupPointRepo)
	if err != nil {
		h.logger(r.Context()).Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	points := make([]pickupPointJSON, len(out.PickupPoints))
	for i, point := range out.PickupPoints {
		points[i] = pickupPointJSON{
			ID:        point.ID,
			Longitude: point.Longitude,
			Latitude:  point.Latitude,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pickupPointsResponse{PickupPoints: points}); err != nil {
		h.logger(r.Context()).Error(err.Error())
	}
}
