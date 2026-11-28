package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/products/domain"

	"go.uber.org/zap"
)

type getProductResponse struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Category        string   `json:"category"`
	PictureURLS     []string `json:"pictureUrls"`
	Price           float64  `json:"price"`
	Discount        uint     `json:"discount"`
	PriceResult     float64  `json:"priceResult"`
	AvailableCount  uint     `json:"availableCount"`
	Description     string   `json:"description"`
	SellerName      string   `json:"sellerName"`
	SellerAvatarURL string   `json:"sellerAvatarUrl"`
	Rating          float64  `json:"rating"`
	ReviewsCount    uint     `json:"reviewsCount"`
	MarksCount      uint     `json:"marksCount"`
	Rating1Count    uint     `json:"rating1Count"`
	Rating2Count    uint     `json:"rating2Count"`
	Rating3Count    uint     `json:"rating3Count"`
	Rating4Count    uint     `json:"rating4Count"`
	Rating5Count    uint     `json:"rating5Count"`
	CanReview       bool     `json:"canReview"`
	WasReview       bool     `json:"wasReview"`
}

func (h *ProductsHandler) Product(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}
	accessToken, err := middleware.ParseAccessTokenPayload(r, h.jwtSecret)
	canReview := false
	var buyerID *int64 = nil
	if err == nil && accessToken != nil {
		canReview = true
		buyerID = &accessToken.UserID
	}

	logger.Debug("Can Review", zap.Bool("CanReview", canReview))
	productByID, err := h.productUsecase.Product(id, buyerID)
	if errors.Is(err, domain.ErrProductIDNotFound) {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {

		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	description := ""
	if productByID.Description != nil {
		description = *productByID.Description
	}
	avatarURL := ""
	if productByID.SellerAvatarURL != nil {
		avatarURL = *productByID.SellerAvatarURL
	}
	out := getProductResponse{ID: productByID.ID, Name: productByID.Name,
		Category: productByID.Category, PictureURLS: productByID.PictureURLS, Price: float64(productByID.Price) / 100,
		Discount: productByID.Discount, PriceResult: float64(productByID.PriceResult) / 100, AvailableCount: productByID.AvailableCount,
		Description: description, SellerName: productByID.SellerName, SellerAvatarURL: avatarURL,
		Rating: productByID.Rating, ReviewsCount: productByID.ReviewsCount, MarksCount: productByID.MarksCount,
		Rating1Count: productByID.Rating1Count, Rating2Count: productByID.Rating2Count, Rating3Count: productByID.Rating3Count,
		Rating4Count: productByID.Rating4Count, Rating5Count: productByID.Rating5Count, CanReview: canReview && productByID.CanReview,
		WasReview: productByID.WasReview,
	}
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(out)
	if err != nil {
		logger.Error(err.Error())
	}
}
