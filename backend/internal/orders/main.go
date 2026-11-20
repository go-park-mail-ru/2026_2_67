package orders

import (
	"fmt"
	"net/http"
	"vibe_market/backend/internal/auth/middleware"
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
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listOrders.ServeHTTP(w, r)
		case http.MethodPost:
			createOrder.ServeHTTP(w, r)
		default:
			w.Header().Set("Allow", fmt.Sprintf("%s, %s", http.MethodGet, http.MethodPost))
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return mux
}
