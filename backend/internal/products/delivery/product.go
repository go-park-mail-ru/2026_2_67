package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"vibe_market/backend/internal/products/domain"
	"vibe_market/backend/internal/products/usecase"
)

type getProductResponse struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	Category       string   `json:"category"`
	PictureURLS    []string `json:"pictureUrls"`
	Price          float64  `json:"price"`
	Discount       uint     `json:"discount"`
	PriceResult    float64  `json:"priceResult"`
	AvailableCount uint     `json:"availableCount"`
	Description    string   `json:"description"`
	Raiting        float64  `json:"raiting"`
	MarksCount     uint     `json:"marksCount"`
	ReviewsCount   uint     `json:"reviewsCount"`
}

func (h *ProductsHandler) Product(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	productByID, err := usecase.Product(id, h.productRepo)
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
	out := getProductResponse{ID: productByID.ID, Name: productByID.Name,
		Category: productByID.Category, PictureURLS: productByID.PictureURLS, Price: productByID.Price,
		Discount: productByID.Discount, PriceResult: productByID.PriceResult, AvailableCount: productByID.AvailableCount,
		Description: description, Raiting: productByID.Raiting, MarksCount: productByID.MarksCount, ReviewsCount: productByID.ReviewsCount,
	}
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(out)
	if err != nil {
		logger.Error(err.Error())
	}
}
