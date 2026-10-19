package buyers

import (
	"net/http"
	"vibe_market/backend/internal/buyers/controller"
	"vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/common"
)

func NewServeMux(buyerRepo domain.BuyerRepo) *http.ServeMux {
	mux := http.NewServeMux()
	buyersHandler := controller.NewBuyersHandler(buyerRepo)

	mux.Handle("/{id}/profile",
		common.RequireHTTPMethodMiddleware(http.MethodGet)(http.HandlerFunc(buyersHandler.GetBuyerProfile)))

	return mux
}
