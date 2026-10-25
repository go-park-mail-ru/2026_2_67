package buyers

import (
	"net/http"
	"vibe_market/backend/internal/buyers/delivery"
	"vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

func NewServeMux(buyerRepo domain.BuyerRepo, baseLogger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	buyersHandler := delivery.NewBuyersHandler(buyerRepo, baseLogger)

	mux.Handle("/{id}/profile",
		common.RequireHTTPMethodMiddleware(http.MethodGet)(http.HandlerFunc(buyersHandler.GetBuyerProfile)))

	return mux
}
