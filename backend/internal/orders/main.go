package orders

import (
	"fmt"
	"net/http"
	"vibe_market/backend/internal/orders/delivery"
	"vibe_market/backend/internal/orders/domain"

	"go.uber.org/zap"
)

func NewServeMux(
	orderRepo domain.OrderRepo, jwtSecret []byte, logger *zap.Logger,
) *http.ServeMux {
	mux := http.NewServeMux()

	ordersHandler := delivery.NewOrdersHandler(orderRepo, jwtSecret, logger)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ordersHandler.ListOrders(w, r)
		case http.MethodPost:
			ordersHandler.CreateOrder(w, r)
		default:
			w.Header().Set("Allow", fmt.Sprintf("%s, %s", http.MethodGet, http.MethodPost))
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return mux
}
