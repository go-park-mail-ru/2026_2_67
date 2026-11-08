package buyers

import (
	"net/http"
	"vibe_market/backend/internal/buyers/delivery"
	"vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

func NewServeMux(
	buyerRepo domain.BuyerRepo,
	basketRepo domain.BasketRepo,
	jwtSecret []byte,
	baseLogger *zap.Logger,
	avatarStorage domain.BuyerAvatarRepo,
) *http.ServeMux {
	mux := http.NewServeMux()
	buyersHandler := delivery.NewBuyersHandler(buyerRepo, basketRepo, jwtSecret, baseLogger, avatarStorage)

	mux.Handle("/{id}/profile", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			buyersHandler.GetBuyerProfile(w, r)
		case http.MethodPatch:
			buyersHandler.PatchBuyerProfile(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.Handle("/{id}/basket",
		common.RequireHTTPMethodMiddleware(http.MethodGet)(http.HandlerFunc(buyersHandler.GetBasket)))

	mux.Handle("/{id}/basket/products",
		common.RequireHTTPMethodMiddleware(http.MethodPost)(http.HandlerFunc(buyersHandler.AddBasketProduct)))

	mux.Handle("/{buyer_id}/basket/products/{product_id}",
		common.RequireHTTPMethodMiddleware(http.MethodPatch)(http.HandlerFunc(buyersHandler.UpdateBasketProduct)))

	mux.Handle("/{id}/basket/preview",
		common.RequireHTTPMethodMiddleware(http.MethodPost)(http.HandlerFunc(buyersHandler.PreviewBasket)))

	return mux
}
