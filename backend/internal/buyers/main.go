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
	basketUsecase := usecase.NewBasketUsecase(basketRepo)
	buyersHandler := delivery.NewBuyersHandler(baseLogger, avatarStorage, buyersUsecase, basketUsecase)

	mux.Handle("/buyers/{id}/profile", common.MultimethodEndpointMiddleware(
		map[string]http.Handler{
			http.MethodGet:   http.HandlerFunc(buyersHandler.Buyer),
			http.MethodPatch: middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(buyersHandler.PatchBuyer)),
		}))

	mux.Handle("/buyers/{id}/basket",
		common.RequireHTTPMethodMiddleware(http.MethodGet)(middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(buyersHandler.Basket))))

	mux.Handle("/buyers/{id}/basket/is_valid",
		common.RequireHTTPMethodMiddleware(http.MethodGet)(middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(buyersHandler.BasketIsValid))))

	mux.Handle("/buyers/{id}/basket/products",
		common.RequireHTTPMethodMiddleware(http.MethodPost)(middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(buyersHandler.PostBasketProduct))))

	mux.Handle("/buyers/{buyer_id}/basket/products/{product_id}",
		common.RequireHTTPMethodMiddleware(http.MethodPatch)(middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(buyersHandler.PatchBasketProduct))))

	mux.Handle("/buyers/{id}/basket/preview",
		common.RequireHTTPMethodMiddleware(http.MethodPost)(middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(buyersHandler.PreviewBasket))))

	return mux
}
