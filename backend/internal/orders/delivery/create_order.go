package delivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/orders/usecase"
)

type createOrderRequest struct {
	Longitude *float64 `json:"longitude"`
	Latitude  *float64 `json:"latitude"`
	Promocode any      `json:"promocode"`
}

type orderValidationResponse struct {
	ErrMessage          string `json:"errMessage"`
	LongitudeErrMessage string `json:"longitudeErrMessage"`
	LatitudeErrMessage  string `json:"latitudeErrMessage"`
	PromocodeErrMessage string `json:"promocodeErrMessage"`
}

type productStockErrorItem struct {
	ProductID  int64  `json:"productId"`
	Name       string `json:"name"`
	Seller     string `json:"seller"`
	ErrMessage string `json:"errMessage"`
}

type orderStockErrorResponse struct {
	ErrMessage  string                  `json:"errMessage"`
	ProductsErr []productStockErrorItem `json:"productsErr"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func (h *OrdersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	logger := h.logger(r.Context())

	authPayload, status, err := h.authenticate(r)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(status)
		return
	}

	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Longitude == nil || req.Latitude == nil {
		logger.Warn("invalid body for create order")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	promocodeStr := ""
	if req.Promocode != nil {
		switch v := req.Promocode.(type) {
		case string:
			promocodeStr = v
		case float64:
			promocodeStr = fmt.Sprintf("%.0f", v)
		}
	}

	out, valErrs, stockErrors, err := usecase.CreateOrderUsecase(r.Context(), usecase.CreateOrderInDTO{
		BuyerID:   authPayload.UserID,
		Longitude: *req.Longitude,
		Latitude:  *req.Latitude,
		Promocode: promocodeStr,
	}, h.orderRepo)

	if valErrs.HasErrors() {
		logger.Warn("order validation errors")
		writeJSON(w, http.StatusUnprocessableEntity, orderValidationResponse{
			ErrMessage:          "Body содержит неверные данные",
			LongitudeErrMessage: valErrs.LongitudeErrMessage,
			LatitudeErrMessage:  valErrs.LatitudeErrMessage,
			PromocodeErrMessage: valErrs.PromocodeErrMessage,
		})
		return
	}

	if len(stockErrors) > 0 {
		logger.Warn("order stock errors")
		items := make([]productStockErrorItem, 0, len(stockErrors))
		for _, se := range stockErrors {
			items = append(items, productStockErrorItem{
				ProductID:  se.ProductID,
				Name:       se.Name,
				Seller:     se.Seller,
				ErrMessage: se.ErrMessage,
			})
		}
		writeJSON(w, http.StatusUnprocessableEntity, orderStockErrorResponse{
			ErrMessage:  "при создании заказа произошла ошибка",
			ProductsErr: items,
		})
		return
	}

	if errors.Is(err, domain.ErrOrderBasketEmpty) {
		logger.Warn(err.Error())
		writeJSON(w, http.StatusUnprocessableEntity, orderStockErrorResponse{
			ErrMessage:  "при создании заказа произошла ошибка",
			ProductsErr: []productStockErrorItem{},
		})
		return
	}

	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, out)
}
