package orders

import (
	"net/http"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/orders/delivery"
	"vibe_market/backend/internal/orders/domain"

	"go.uber.org/zap"
)

func NewServeMux(
	orderRepo domain.OrderRepo, jwtSecret []byte, logger *zap.Logger,
) *http.ServeMux {
	mux := http.NewServeMux()

	ordersHandler := delivery.NewOrdersHandler(orderRepo, jwtSecret, logger)
	listOrders := middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(ordersHandler.ListOrders))
	createOrder := middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(ordersHandler.CreateOrder))
	handler := common.MultimethodEndpointMiddleware(map[string]http.Handler{
		http.MethodGet:  listOrders,
		http.MethodPost: createOrder,
	})
	mux.Handle("/orders", handler)
	mux.Handle("/orders/{$}", handler)

	return mux
}
