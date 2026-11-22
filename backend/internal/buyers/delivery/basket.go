package delivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/buyers/usecase"
)

type errorPostBasketProductNotFoundResponse struct {
	ErrMessage string `json:"errMessage"`
}

type errorPatchBasketProductNotFoundResponse struct {
	ErrMessage string `json:"errMessage"`
}

type errorBasketIsValidProduct struct {
	ProductID         int64 `json:"id"`
	NewAvailableCount uint  `json:"newAvailableCount"`
}

type errorBasketIsValidResponse struct {
	Products []errorBasketIsValidProduct `json:"products"`
}

type postPreviewBasketValidationResponse struct {
	ErrMessage          string `json:"errMessage"`
	LongitudeErrMessage string `json:"longitudeErrMessage"`
	LatitudeErrMessage  string `json:"latitudeErrMessage"`
	PromocodeErrMessage string `json:"promocodeErrMessage"`
}

type postPreviewBasketProductError struct {
	ProductID  int64  `json:"productId"`
	Name       string `json:"name"`
	Seller     string `json:"seller"`
	ErrMessage string `json:"errMessage"`
}

type postPreviewBasketStockResponse struct {
	ErrMessage  string                          `json:"errMessage"`
	ProductsErr []postPreviewBasketProductError `json:"productsErr"`
}

type postBasketProductRequest struct {
	ProductID *int64 `json:"productId"`
	Count     *int   `json:"count"`
}

type patchBasketProductRequest struct {
	Count *int `json:"count"`
}

type postPreviewBasketRequest struct {
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

func (h *BuyersHandler) Basket(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	buyerID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authPayload, _ := r.Context().Value(middleware.AccessTokenPayloadKey).(*authDomain.AccessTokenPayload)
	if authPayload == nil || authPayload.UserID != buyerID {
		logger.Warn("идентификатор пользователя не совпадает")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	out, err := h.basketUsecase.Basket(buyerID)
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

func (h *BuyersHandler) BasketIsValid(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	buyerID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authPayload, _ := r.Context().Value(middleware.AccessTokenPayloadKey).(*authDomain.AccessTokenPayload)
	if authPayload == nil || authPayload.UserID != buyerID {
		logger.Warn("идентификатор пользователя не совпадает")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	out, err := h.basketUsecase.IsValid(buyerID)
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
	products := make([]errorBasketIsValidProduct, 0, len(out.Products))
	for i := range out.Products {
		products = append(products, errorBasketIsValidProduct{out.Products[i], out.NewAvailableCounts[i]})
	}
	if len(out.Products) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, errorBasketIsValidResponse{
			Products: products,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BuyersHandler) PostBasketProduct(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	buyerID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authPayload, _ := r.Context().Value(middleware.AccessTokenPayloadKey).(*authDomain.AccessTokenPayload)
	if authPayload == nil || authPayload.UserID != buyerID {
		logger.Warn("идентификатор пользователя не совпадает")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var req postBasketProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID == nil || req.Count == nil {
		logger.Warn("некорректное тело запроса на добавление товара")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.basketUsecase.AddProduct(buyerID, *req.ProductID, *req.Count)
	if errors.Is(err, domain.ErrBuyerNotFound) {
		logger.Warn(err.Error())
		writeJSON(w, http.StatusNotFound, errorPostBasketProductNotFoundResponse{
			ErrMessage: fmt.Sprintf("не найден пользователь с id = %d", buyerID),
		})
		return
	}
	if errors.Is(err, domain.ErrProductNotFound) {
		logger.Warn(err.Error())
		writeJSON(w, http.StatusNotFound, errorPostBasketProductNotFoundResponse{
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

func (h *BuyersHandler) PatchBasketProduct(w http.ResponseWriter, r *http.Request) {
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

	authPayload, _ := r.Context().Value(middleware.AccessTokenPayloadKey).(*authDomain.AccessTokenPayload)
	if authPayload == nil || authPayload.UserID != buyerID {
		logger.Warn("идентификатор пользователя не совпадает")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var req patchBasketProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Count == nil {
		logger.Warn("некорректное тело запроса на изменение товара")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.basketUsecase.UpdateProductCount(buyerID, productID, *req.Count)
	if errors.Is(err, domain.ErrBuyerNotFound) {
		logger.Warn(err.Error())
		writeJSON(w, http.StatusNotFound, errorPatchBasketProductNotFoundResponse{
			ErrMessage: fmt.Sprintf("не найден пользователь с id = %d", buyerID),
		})
		return
	}
	if errors.Is(err, domain.ErrProductNotFoundInBasket) || errors.Is(err, domain.ErrProductNotFound) {
		logger.Warn(err.Error())
		writeJSON(w, http.StatusNotFound, errorPatchBasketProductNotFoundResponse{
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

	authPayload, _ := r.Context().Value(middleware.AccessTokenPayloadKey).(*authDomain.AccessTokenPayload)
	if authPayload == nil || authPayload.UserID != buyerID {
		logger.Warn("идентификатор пользователя не совпадает")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var req postPreviewBasketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Longitude == nil || req.Latitude == nil {
		logger.Warn("некорректное тело запроса на предварительный просмотр корзины")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	promocode := ""
	if req.Promocode != nil {
		promocode = *req.Promocode
	}

	out, valErrs, stockErrors, err := h.basketUsecase.PreviewBasket(buyerID, usecase.PreviewBasketInDTO{
		Longitude: *req.Longitude,
		Latitude:  *req.Latitude,
		Promocode: promocode,
	})

	if valErrs.HasErrors() {
		logger.Warn("ошибки валидации предварительного просмотра корзины")
		writeJSON(w, http.StatusUnprocessableEntity, postPreviewBasketValidationResponse{
			ErrMessage:          "тело запроса содержит неверные данные",
			LongitudeErrMessage: valErrs.LongitudeErrMessage,
			LatitudeErrMessage:  valErrs.LatitudeErrMessage,
			PromocodeErrMessage: valErrs.PromocodeErrMessage,
		})
		return
	}

	if len(stockErrors) > 0 {
		logger.Warn("ошибки доступного остатка товаров")
		stockItems := make([]postPreviewBasketProductError, 0, len(stockErrors))
		for _, se := range stockErrors {
			stockItems = append(stockItems, postPreviewBasketProductError{
				ProductID:  se.ProductID,
				Name:       se.Name,
				Seller:     se.Seller,
				ErrMessage: se.ErrMessage,
			})
		}
		writeJSON(w, http.StatusUnprocessableEntity, postPreviewBasketStockResponse{
			ErrMessage:  "при создании заказа произошла ошибка",
			ProductsErr: stockItems,
		})
		return
	}

	if errors.Is(err, domain.ErrBasketEmpty) {
		logger.Warn(err.Error())
		writeJSON(w, http.StatusUnprocessableEntity, postPreviewBasketStockResponse{
			ErrMessage:  "при создании заказа произошла ошибка",
			ProductsErr: []postPreviewBasketProductError{},
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
