package products

import (
	"net/http"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/products/delivery"
	"vibe_market/backend/internal/products/domain"

	"go.uber.org/zap"
)

func NewServeMux(productsRepo domain.ProductRepo, logger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	productsHandler := delivery.NewProductsHandler(productsRepo, logger)
	mux.Handle("/", common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(productsHandler.Products)))

	return mux
}
