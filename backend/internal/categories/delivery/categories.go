package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"vibe_market/backend/internal/categories/domain"
	"vibe_market/backend/internal/categories/usecase"
)

type categoryJSON struct {
	ID         int64  `json:"id"`
	Category   string `json:"category"`
	PictureURL string `json:"pictureUrl"`
}

type categoriesResponse struct {
	Categories []categoryJSON `json:"productCategories"`
}

func (h *CategoriesHandler) Categories(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	var params domain.CategoriesSearchParams
	params.SortBy = domain.SortByType(r.URL.Query().Get("sortBy"))
	limitRaw := r.URL.Query().Get("limit")
	if limitRaw != "" {
		limit, err := strconv.ParseUint(limitRaw, 10, 32)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			params.Limit = uint(limit)
		}
	}

	categories, err := usecase.Categories(params, h.categoryRepo)
	if errors.Is(err, domain.ErrInvalidSortBy) {
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
