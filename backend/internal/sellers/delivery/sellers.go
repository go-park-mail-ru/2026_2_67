package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"
	"vibe_market/backend/internal/sellers/domain"
)

type sellerJSON struct {
	UserID    int64  `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

type sellersResponse struct {
	Sellers []sellerJSON `json:"sellers"`
}

func (h *SellersHandler) Sellers(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	var searchParams domain.SellerSearchParams
	limitRaw := r.URL.Query().Get("limit")
	if limitRaw != "" {
		limit, err := strconv.ParseUint(limitRaw, 10, 32)
		if err != nil {
			searchParams.Limit = 0
			logger.Warn(err.Error())
		} else {
			searchParams.Limit = uint(limit)
		}
	}
	searchParams.Category = r.URL.Query()["category"]
	searchParams.Query = r.URL.Query().Get("query")
	offsetRaw := r.URL.Query().Get("offset")
	if offsetRaw != "" {
		offset, err := strconv.ParseUint(offsetRaw, 10, 32)
		if err != nil {
			searchParams.Offset = 0
			logger.Warn(err.Error())
		} else {
			searchParams.Offset = uint(offset)
		}
	}
	priceMinRaw := r.URL.Query().Get("priceMin")
	if priceMinRaw != "" {
		priceMin, err := strconv.ParseFloat(priceMinRaw, 64)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			searchParams.PriceMin = priceMin
		}
	}
	priceMaxRaw := r.URL.Query().Get("priceMax")
	if priceMaxRaw != "" {
		priceMax, err := strconv.ParseFloat(priceMaxRaw, 64)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			searchParams.PriceMax = &priceMax
		}
	}
	if ratingMinRaw := r.URL.Query().Get("ratingMin"); ratingMinRaw != "" {
		ratingMin, err := strconv.ParseFloat(ratingMinRaw, 64)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			searchParams.RatingMin = ratingMin
		}
	}
	if ratingMaxRaw := r.URL.Query().Get("ratingMax"); ratingMaxRaw != "" {
		ratingMax, err := strconv.ParseFloat(ratingMaxRaw, 64)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			searchParams.RatingMax = &ratingMax
		}
	}

	sellers, err := h.usecase.Sellers(searchParams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(err.Error())
		return
	}
	response := sellersResponse{make([]sellerJSON, len(sellers.Sellers))}
	for i, seller := range sellers.Sellers {
		response.Sellers[i].UserID = seller.UserID
		response.Sellers[i].Name = seller.Name
		if seller.AvatarURL != nil {
			response.Sellers[i].AvatarURL = *seller.AvatarURL
		}
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	err = encoder.Encode(response)
	if err != nil {
		logger.Error(err.Error())
	}
}
