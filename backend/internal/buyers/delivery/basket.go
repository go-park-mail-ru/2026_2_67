package delivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/buyers/usecase"
)

type errorMessageResponse struct {
	ErrMessage string `json:"errMessage"`
}

type previewValidationResponse struct {
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

type previewStockErrorResponse struct {
	ErrMessage  string                  `json:"errMessage"`
	ProductsErr []productStockErrorItem `json:"productsErr"`
}

type addProductRequest struct {
	ProductID *int64 `json:"productId"`
	Count     *int   `json:"count"`
}

type updateProductRequest struct {
	Count *int `json:"count"`
}

type previewRequest struct {
	Longitude *float64 `json:"longitude"`
	Latitude  *float64 `json:"latitude"`
	Promocode *string  `json:"promocode"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func (h *BuyersHandler) GetBasket(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	buyerID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authPayload, status, err := h.authenticate(r)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(status)
		return
	}

	if authPayload.UserID != buyerID {
		logger.Warn("user id mismatch")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	out, err := usecase.GetBasketUsecase(buyerID, h.basketRepo)
	if errors.Is(err, domain.ErrBuyerNotFound) {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, out)
}

func (h *BuyersHandler) AddBasketProduct(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	buyerID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authPayload, status, err := h.authenticate(r)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(status)
		return
	}

	if authPayload.UserID != buyerID {
		logger.Warn("user id mismatch")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var req addProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID == nil || req.Count == nil {
		logger.Warn("invalid body for add product")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = usecase.AddProductUsecase(buyerID, *req.ProductID, *req.Count, h.basketRepo)
	if errors.Is(err, domain.ErrBuyerNotFound) {
		logger.Warn(err.Error())
		writeJSON(w, http.StatusNotFound, errorMessageResponse{
			ErrMessage: fmt.Sprintf("не найден пользователь с id = %d", buyerID),
		})
		return
	}
	if errors.Is(err, domain.ErrProductNotFound) {
		logger.Warn(err.Error())
		writeJSON(w, http.StatusNotFound, errorMessageResponse{
			ErrMessage: fmt.Sprintf("не найден товар с id = %d", *req.ProductID),
		})
		return
	}
	if errors.Is(err, domain.ErrCountInvalid) || errors.Is(err, domain.ErrCountExceeded) {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BuyersHandler) UpdateBasketProduct(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	buyerID, err := strconv.ParseInt(r.PathValue("buyer_id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	productID, err := strconv.ParseInt(r.PathValue("product_id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authPayload, status, err := h.authenticate(r)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(status)
		return
	}

	if authPayload.UserID != buyerID {
		logger.Warn("user id mismatch")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var req updateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Count == nil {
		logger.Warn("invalid body for update product")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = usecase.UpdateProductCountUsecase(buyerID, productID, *req.Count, h.basketRepo)
	if errors.Is(err, domain.ErrBuyerNotFound) {
		logger.Warn(err.Error())
		writeJSON(w, http.StatusNotFound, errorMessageResponse{
			ErrMessage: fmt.Sprintf("не найден пользователь с id = %d", buyerID),
		})
		return
	}
	if errors.Is(err, domain.ErrProductNotFoundInBasket) || errors.Is(err, domain.ErrProductNotFound) {
		logger.Warn(err.Error())
		writeJSON(w, http.StatusNotFound, errorMessageResponse{
			ErrMessage: fmt.Sprintf("не найден товар с id = %d", productID),
		})
		return
	}
	if errors.Is(err, domain.ErrCountInvalid) || errors.Is(err, domain.ErrCountExceeded) {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BuyersHandler) PreviewBasket(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	buyerID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authPayload, status, err := h.authenticate(r)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(status)
		return
	}

	if authPayload.UserID != buyerID {
		logger.Warn("user id mismatch")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var req previewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Longitude == nil || req.Latitude == nil {
		logger.Warn("invalid body for preview basket")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	promocode := ""
	if req.Promocode != nil {
		promocode = *req.Promocode
	}

	out, valErrs, stockErrors, err := usecase.PreviewBasketUsecase(buyerID, usecase.PreviewInDTO{
		Longitude: *req.Longitude,
		Latitude:  *req.Latitude,
		Promocode: promocode,
	}, h.basketRepo)

	if valErrs.HasErrors() {
		logger.Warn("preview validation errors")
		writeJSON(w, http.StatusUnprocessableEntity, previewValidationResponse{
			ErrMessage:          "Body содержит неверные данные",
			LongitudeErrMessage: valErrs.LongitudeErrMessage,
			LatitudeErrMessage:  valErrs.LatitudeErrMessage,
			PromocodeErrMessage: valErrs.PromocodeErrMessage,
		})
		return
	}

	if len(stockErrors) > 0 {
		logger.Warn("preview stock errors")
		stockItems := make([]productStockErrorItem, 0, len(stockErrors))
		for _, se := range stockErrors {
			stockItems = append(stockItems, productStockErrorItem{
				ProductID:  se.ProductID,
				Name:       se.Name,
				Seller:     se.Seller,
				ErrMessage: se.ErrMessage,
			})
		}
		writeJSON(w, http.StatusUnprocessableEntity, previewStockErrorResponse{
			ErrMessage:  "при создании заказа произошла ошибка",
			ProductsErr: stockItems,
		})
		return
	}

	if errors.Is(err, domain.ErrBasketEmpty) {
		logger.Warn(err.Error())
		writeJSON(w, http.StatusUnprocessableEntity, previewStockErrorResponse{
			ErrMessage:  "при создании заказа произошла ошибка",
			ProductsErr: []productStockErrorItem{},
		})
		return
	}

	if errors.Is(err, domain.ErrBuyerNotFound) {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, out)
}
