package delivery

import (
	"net/http"
	"vibe_market/backend/internal/orders/usecase"
)

func (h *OrdersHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	payload, status, err := h.authenticate(r)
	if err != nil {
		h.logger(r.Context()).Warn(err.Error())
		w.WriteHeader(status)
		return
	}
	out, err := usecase.ListOrdersUsecase(r.Context(), payload.UserID, h.orderRepo)
	if err != nil {
		h.logger(r.Context()).Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
