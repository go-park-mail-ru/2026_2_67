package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/middleware"
	orderDomain "vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/products/domain"
	"vibe_market/backend/internal/products/usecase"
)

type postProductReviewRequest struct {
	Review *string `json:"review"`
	Rating *int    `json:"rating"`
}

type postProductReview422Response struct {
	ErrMessage string `json:"errMessage"`
}

func (h *ProductsHandler) PostProductReview(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())
	accessToken := r.Context().Value(middleware.AccessTokenPayloadKey).(*authDomain.AccessTokenPayload)

	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	decoder := json.NewDecoder(r.Body)

	defer r.Body.Close()

	var request postProductReviewRequest
	err = decoder.Decode(&request)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if request.Rating == nil {
		logger.Warn("не указан rating")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = h.productReviewUsecase.PostProductReview(usecase.PostProductReviewInDTO{
		ProductID: productID, BuyerID: accessToken.UserID, Review: request.Review, Rating: *request.Rating})

	if errors.Is(err, orderDomain.ErrOrdersNotFound) || errors.Is(err, domain.ErrProductReviewRatingInvalid) {
		encoder := json.NewEncoder(w)
		w.WriteHeader(http.StatusUnprocessableEntity)
		err = encoder.Encode(postProductReview422Response{
			ErrMessage: err.Error(),
		})
		if err != nil {
			logger.Error(err.Error())
		}
		return
	}
	if errors.Is(err, domain.ErrProductReviewAlreadyExists) {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
