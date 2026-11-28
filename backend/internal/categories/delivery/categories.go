package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"vibe_market/backend/internal/categories/domain"
)

type categoryJSON struct {
	ID         int64  `json:"id"`
	Category   string `json:"category"`
	PictureURL string `json:"pictureUrl"`
}

type categoriesResponse struct {
	Categories []categoryJSON `json:"categories"`
}

func (h *CategoriesHandler) Categories(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	var params domain.CategoriesSearchParams
	params.SortBy = domain.SortByType(r.URL.Query().Get("sortBy"))
	params.Seller = r.URL.Query()["seller"]
	params.Query = r.URL.Query().Get("query")
	limitRaw := r.URL.Query().Get("limit")
	if limitRaw != "" {
		limit, err := strconv.ParseUint(limitRaw, 10, 32)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			params.Limit = uint(limit)
		}
	}
	if priceMinRaw := r.URL.Query().Get("priceMin"); priceMinRaw != "" {
		priceMin, err := strconv.ParseFloat(priceMinRaw, 64)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			params.PriceMin = priceMin
		}
	}
	if priceMaxRaw := r.URL.Query().Get("priceMax"); priceMaxRaw != "" {
		priceMax, err := strconv.ParseFloat(priceMaxRaw, 64)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			params.PriceMax = &priceMax
		}
	}
	if ratingMinRaw := r.URL.Query().Get("ratingMin"); ratingMinRaw != "" {
		ratingMin, err := strconv.ParseFloat(ratingMinRaw, 64)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			params.RatingMin = ratingMin
		}
	}
	if ratingMaxRaw := r.URL.Query().Get("ratingMax"); ratingMaxRaw != "" {
		ratingMax, err := strconv.ParseFloat(ratingMaxRaw, 64)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			params.RatingMax = &ratingMax
		}
	}

	categories, err := h.usecase.Categories(params)
	if errors.Is(err, domain.ErrSortByInvalid) {
		logger.Warn(err.Error())
	} else if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := categoriesResponse{Categories: make([]categoryJSON, len(categories.Categories))}
	for i, category := range categories.Categories {
		response.Categories[i].ID = category.ID
		response.Categories[i].Category = category.Category
		if category.PictureURL != nil {
			response.Categories[i].PictureURL = *category.PictureURL
		}
	}

	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(response)
	if err != nil {
		logger.Error(err.Error())
	}
}
