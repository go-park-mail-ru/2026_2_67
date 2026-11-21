package buyers

import (
	"net/http"
	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/buyers/delivery"
	"vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/buyers/usecase"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

func NewServeMux(
	buyerRepo domain.BuyerRepo,
	basketRepo domain.BasketRepo,
	userRepo authDomain.UserRepo,
	jwtSecret []byte,
	baseLogger *zap.Logger,
	avatarStorage domain.BuyerAvatarRepo,
) *http.ServeMux {
	mux := http.NewServeMux()
	buyersUsecase := usecase.NewBuyerUsecase(buyerRepo, avatarStorage, userRepo)
	buyersHandler := delivery.NewBuyersHandler(buyerRepo, basketRepo, userRepo, baseLogger, avatarStorage, buyersUsecase)

	mux.Handle("/{id}/profile", common.MultimethodEndpointMiddleware(
		map[string]http.Handler{
			http.MethodGet:   http.HandlerFunc(buyersHandler.GetBuyerProfile),
			http.MethodPatch: middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(buyersHandler.PatchBuyerProfile)),
		}))

	mux.Handle("/{id}/basket",
		common.RequireHTTPMethodMiddleware(http.MethodGet)(middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(buyersHandler.GetBasket))))

	mux.Handle("/{id}/basket/products",
		common.RequireHTTPMethodMiddleware(http.MethodPost)(middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(buyersHandler.AddBasketProduct))))

	mux.Handle("/{buyer_id}/basket/products/{product_id}",
		common.RequireHTTPMethodMiddleware(http.MethodPatch)(middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(buyersHandler.UpdateBasketProduct))))

	mux.Handle("/{id}/basket/preview",
		common.RequireHTTPMethodMiddleware(http.MethodPost)(middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(buyersHandler.PreviewBasket))))

	return mux
}
