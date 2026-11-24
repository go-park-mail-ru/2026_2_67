package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (h *CategoriesHandler) Search(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())
	limit := uint(0)
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.ParseUint(rawLimit, 10, 32)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			limit = uint(parsedLimit)
		}
	}

	result, err := h.usecase.Search(r.URL.Query().Get("query"), limit)
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := categoriesResponse{Categories: make([]categoryJSON, 0, len(result.Categories))}
	for _, category := range result.Categories {
		item := categoryJSON{ID: category.ID, Category: category.Category}
		if category.PictureURL != nil {
			item.PictureURL = *category.PictureURL
		}
		response.Categories = append(response.Categories, item)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(err.Error())
	}
}
