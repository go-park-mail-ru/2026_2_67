package sellers

import (
	"net/http"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/sellers/delivery"
	"vibe_market/backend/internal/sellers/domain"

	"go.uber.org/zap"
)

func NewServeMux(sellerRepo domain.SellerRepo, baseLogger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	sellerHandler := delivery.NewSellersHandler(sellerRepo, baseLogger)

	mux.Handle("/",
		common.RequireHTTPMethodMiddleware(http.MethodGet)(http.HandlerFunc(sellerHandler.Sellers)))

	return mux
}
