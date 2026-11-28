package orders

import (
	"net/http"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/orders/delivery"
	"vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/orders/usecase"

	"go.uber.org/zap"
)

func NewServeMux(orderRepo domain.OrderRepo, jwtSecret []byte, onlineConf delivery.OnlineConfClient, logger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	orderUsecase := usecase.NewOrderUsecase(orderRepo)
	ordersHandler := delivery.NewOrdersHandler(orderRepo, orderUsecase, onlineConf, logger)
	listOrders := middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(ordersHandler.Orders))
	createOrder := middleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(ordersHandler.CreateOrder))
	handler := common.MultimethodEndpointMiddleware(map[string]http.Handler{
		http.MethodGet:  listOrders,
		http.MethodPost: createOrder,
	})
	mux.Handle("/", handler)

	return mux
}
