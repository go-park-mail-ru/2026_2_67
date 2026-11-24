package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type searchSellerResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

type searchResponse struct {
	Sellers []searchSellerResponse `json:"sellers"`
}

func (h *SellersHandler) Search(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())
	var limit uint
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

	response := searchResponse{Sellers: make([]searchSellerResponse, 0, len(result.Sellers))}
	for _, seller := range result.Sellers {
		item := searchSellerResponse{ID: seller.UserID, Name: seller.Name}
		if seller.AvatarURL != nil {
			item.AvatarURL = *seller.AvatarURL
		}
		response.Sellers = append(response.Sellers, item)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(err.Error())
	}
}
