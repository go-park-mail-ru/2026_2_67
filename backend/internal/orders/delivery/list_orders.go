package delivery

import (
	"net/http"
	"time"

	"vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/orders/usecase"
)

type listOrdersResponse struct {
	Orders []orderResponse `json:"orders"`
}

type pickupPointResponse struct {
	ID        int64   `json:"id"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
}

type productResponse struct {
	ProductID       int64    `json:"productId"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Category        string   `json:"category"`
	SellerName      string   `json:"sellerName"`
	SellerAvatarURL string   `json:"sellerAvatarUrl"`
	PictureURLs     []string `json:"pictureUrls"`
	Count           int      `json:"count"`
	UnitPrice       float64  `json:"unitPrice"`
	TotalPrice      float64  `json:"totalPrice"`
}

type orderResponse struct {
	ID          int64               `json:"id"`
	Number      int64               `json:"number"`
	Status      string              `json:"status"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
	PickupPoint pickupPointResponse `json:"pickupPoint"`
	Promocode   *string             `json:"promocode"`
	Products    []productResponse   `json:"products"`
	TotalCount  int64               `json:"totalCount"`
	TotalPrice  float64             `json:"totalPrice"`
}

func (h *OrdersHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	payload, status, err := h.authenticate(r)
	if err != nil {
		h.logger(r.Context()).Warn(err.Error())
		w.WriteHeader(status)
		return
	}
	orders, err := usecase.ListOrders(r.Context(), payload.UserID, h.orderRepo)
	if err != nil {
		h.logger(r.Context()).Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, makeListOrdersResponse(orders))
}

func makeListOrdersResponse(orders []domain.OrderDetails) listOrdersResponse {
	response := listOrdersResponse{Orders: make([]orderResponse, 0, len(orders))}
	n := len(orders)
	for i, order := range orders {
		item := orderResponse{
			ID:          order.ID,
			Number:      int64(n - i + 1),
			Status:      order.Status.Name,
			CreatedAt:   order.CreatedAt.UTC(),
			UpdatedAt:   order.UpdatedAt.UTC(),
			PickupPoint: pickupPointResponse(order.PickupPoint),
			Products:    make([]productResponse, 0, len(order.Products)),
		}
		if order.Promocode != nil {
			code := order.Promocode.Code
			item.Promocode = &code
		}
		for _, product := range order.Products {
			description := ""
			if product.Description != nil {
				description = *product.Description
			}
			totalPrice := float64(product.Count) * float64(product.Price)
			item.Products = append(item.Products, productResponse{
				ProductID:       product.ProductID,
				Name:            product.Name,
				Description:     description,
				Category:        product.Category.Name,
				SellerName:      product.Seller.Name,
				SellerAvatarURL: product.Seller.AvatarURL,
				PictureURLs:     append([]string{}, product.PictureURLs...),
				Count:           product.Count,
				UnitPrice:       float64(product.Price),
				TotalPrice:      totalPrice,
			})
			item.TotalCount += int64(product.Count)
			item.TotalPrice += totalPrice
		}
		response.Orders = append(response.Orders, item)
	}
	return response
}
