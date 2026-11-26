package products

import (
	"net/http"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/common"
	ordersDomain "vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/products/delivery"
	"vibe_market/backend/internal/products/domain"
	"vibe_market/backend/internal/products/usecase"

	"go.uber.org/zap"
)

func NewServeMux(productRepo domain.ProductRepo, productReviewRepo domain.ProductReviewRepo, orderRepo ordersDomain.OrderRepo, logger *zap.Logger, jwtSecret []byte) *http.ServeMux {
	mux := http.NewServeMux()

	productUsecase := usecase.NewProductUsecase(productRepo)
	productReviewUsecase := usecase.NewProductReviewUsecase(productReviewRepo, orderRepo)

	productsHandler := delivery.NewProductsHandler(productRepo, productUsecase, productReviewUsecase, logger, jwtSecret)
	mux.Handle("/search", common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(productsHandler.Search)))
	mux.Handle("/", common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(productsHandler.Products)))

	mux.Handle("/{id}", common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(productsHandler.Product),
	))
	mux.Handle("/{id}/reviews", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			productsHandler.GetProductReviews(w, r)
		case http.MethodPost:
			middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(productsHandler.PostProductReview)).ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/{product_id}/reviews/{buyer_id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			productsHandler.GetProductReview(w, r)
		case http.MethodPatch:
			middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(productsHandler.PatchProductReview)).ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	return mux
}
