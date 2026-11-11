package pickup_points

import (
	"net/http"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/pickup_points/delivery"
	"vibe_market/backend/internal/pickup_points/domain"

	"go.uber.org/zap"
)

func NewServeMux(repo domain.PickupPointRepo, logger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	handler := delivery.NewPickupPointsHandler(repo, logger)
	mux.Handle("/", common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(handler.PickupPoints)))
	return mux
}
