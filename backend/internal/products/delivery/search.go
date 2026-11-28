package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type searchProductResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	PictureURL string `json:"pictureUrl"`
	SellerName string `json:"sellerName"`
}

type searchResponse struct {
	Products []searchProductResponse `json:"products"`
}

func (h *ProductsHandler) Search(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.productUsecase.Search(r.URL.Query().Get("query"), limit)
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := searchResponse{Products: make([]searchProductResponse, 0, len(result.Products))}
	for _, product := range result.Products {
		response.Products = append(response.Products, searchProductResponse{
			ID: product.ID, Name: product.Name, PictureURL: product.PictureURL, SellerName: product.SellerName,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(err.Error())
	}
}
