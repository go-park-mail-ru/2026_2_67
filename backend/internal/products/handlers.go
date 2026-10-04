// Package products для работы с товарами
package products

import (
	"encoding/json"
	"net/http"
)

type ProductsHandler struct {
	storage Storage
}

func MakeProductsHandler(storage Storage) ProductsHandler {
	return ProductsHandler{storage: storage}
}

type ProductResponse struct {
	ProductName         string   `json:"productName"`
	ProductPictureURLs  []string `json:"productPictureUrls"`
	ProductPrice        int      `json:"productPrice"`
	ProductRating       float64  `json:"productRating"`
	ProductReviewsCount int      `json:"productReviewsCount"`
}

// GetProductsHandler реализует роутер GET /api/v1/products
func (h *ProductsHandler) GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	products, err := h.storage.SelectAllProducts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		pictureURLs := p.PictureURLs
		if pictureURLs == nil {
			pictureURLs = []string{}
		}

		response = append(response, ProductResponse{
			ProductName:         p.Name,
			ProductPictureURLs:  pictureURLs,
			ProductPrice:        p.Price,
			ProductRating:       p.Rating,
			ProductReviewsCount: p.ReviewsCount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
