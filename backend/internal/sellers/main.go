package sellers

import (
	"net/http"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/sellers/delivery"
	"vibe_market/backend/internal/sellers/domain"
	"vibe_market/backend/internal/sellers/usecase"

	"go.uber.org/zap"
)

func NewServeMux(sellerRepo domain.SellerRepo, baseLogger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	sellerUsecase := usecase.NewSellerUsecase(sellerRepo)
	sellersHandler := delivery.NewSellersHandler(sellerRepo, sellerUsecase, baseLogger)

	mux.Handle("/sellers",
		common.RequireHTTPMethodMiddleware(http.MethodGet)(http.HandlerFunc(sellersHandler.Sellers)))

	return mux
}
