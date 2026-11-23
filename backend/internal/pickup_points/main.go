package pickup_points

import (
	"net/http"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/pickup_points/delivery"
	"vibe_market/backend/internal/pickup_points/domain"
	"vibe_market/backend/internal/pickup_points/usecase"

	"go.uber.org/zap"
)

func NewServeMux(repo domain.PickupPointRepo, logger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	usecase := usecase.NewPickupPointUsecase(repo)
	handler := delivery.NewPickupPointsHandler(repo, usecase, logger)
	mux.Handle("/pickup_points", common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(handler.PickupPoints)))
	return mux
}
