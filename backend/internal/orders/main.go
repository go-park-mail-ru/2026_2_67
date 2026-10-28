package orders

import (
	"net/http"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/orders/delivery"
	"vibe_market/backend/internal/orders/domain"

	"go.uber.org/zap"
)

func NewServeMux(
	orderRepo domain.OrderRepo,
	jwtSecret []byte,
	logger *zap.Logger,
) *http.ServeMux {
	mux := http.NewServeMux()

	ordersHandler := delivery.NewOrdersHandler(orderRepo, jwtSecret, logger)
	mux.Handle("/", common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(ordersHandler.CreateOrder)))

	return mux
}
