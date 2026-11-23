package delivery

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	authDomain "vibe_market/backend/internal/auth/domain"
	authMiddleware "vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/orders/domain"
)

type ordersResponse struct {
	Orders     []orderResponse `json:"orders"`
	NextCursor string          `json:"nextCursor"`
}

type orderCursorJSON struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        int64     `json:"id"`
}

type pickupPointResponse struct {
	ID        int64   `json:"id"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
}

type productResponse struct {
	ProductID       int64    `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Category        string   `json:"category"`
	SellerName      string   `json:"sellerName"`
	SellerAvatarURL string   `json:"sellerAvatarUrl"`
	PictureURLs     []string `json:"pictureUrls"`
	Count           uint     `json:"count"`
	UnitPrice       float64  `json:"unitPrice"`
	TotalPrice      float64  `json:"totalPrice"`
}

type orderResponse struct {
	ID          int64               `json:"id"`
	Status      string              `json:"status"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
	PickupPoint pickupPointResponse `json:"pickupPoint"`
	Promocode   string              `json:"promocode"`
	Products    []productResponse   `json:"products"`
	TotalCount  uint                `json:"totalCount"`
	TotalPrice  float64             `json:"totalPrice"`
}

func (h *OrdersHandler) Orders(w http.ResponseWriter, r *http.Request) {
	payload, _ := r.Context().Value(authMiddleware.AccessTokenPayloadKey).(*authDomain.AccessTokenPayload)
	if payload == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	params, err := parseOrdersParams(r)
	if err != nil {
		writeJSON(w, http.StatusOK, ordersResponse{})
		return
	}
	page, err := h.usecase.Orders(r.Context(), payload.UserID, params)
	if err != nil {
		h.logger(r.Context()).Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := makeOrdersResponse(page.Orders)
	if page.NextCursor != nil {
		response.NextCursor = encodeCursor(*page.NextCursor)
	}
	writeJSON(w, http.StatusOK, response)
}

func makeOrdersResponse(orders []domain.OrderView) ordersResponse {
	response := ordersResponse{Orders: make([]orderResponse, 0, len(orders))}
	for _, order := range orders {
		item := orderResponse{
			ID:        order.ID,
			Status:    order.Status,
			CreatedAt: order.CreatedAt.UTC(),
			UpdatedAt: order.UpdatedAt.UTC(),
			PickupPoint: pickupPointResponse{
				ID:        order.PickupPoint.ID,
				Longitude: order.PickupPoint.Longitude,
				Latitude:  order.PickupPoint.Latitude,
				StartTime: order.PickupPoint.StartTime.Format("15:04:05"),
				EndTime:   order.PickupPoint.EndTime.Format("15:04:05"),
			},
			Products: make([]productResponse, 0, len(order.Products)),
		}
		if order.Promocode != nil {
			item.Promocode = *order.Promocode
		}
		for _, product := range order.Products {
			unitPrice := float64(product.Price) / 100
			totalPrice := float64(product.TotalPrice) / 100
			avatarURL := ""
			if product.Seller.AvatarURL != nil {
				avatarURL = *product.Seller.AvatarURL
			}
			item.Products = append(item.Products, productResponse{
				ProductID:       product.ProductID,
				Name:            product.Name,
				Description:     product.Description,
				Category:        product.Category,
				SellerName:      product.Seller.Name,
				SellerAvatarURL: avatarURL,
				PictureURLs:     append([]string{}, product.PictureURLS...),
				Count:           product.Count,
				UnitPrice:       unitPrice,
				TotalPrice:      totalPrice,
			})
		}
		item.TotalCount = order.TotalCount
		item.TotalPrice = float64(order.TotalPrice) / 100
		response.Orders = append(response.Orders, item)
	}
	return response
}

func parseOrdersParams(r *http.Request) (domain.OrdersParams, error) {
	query := r.URL.Query()
	params := domain.OrdersParams{Status: query.Get("status"), SortBy: domain.OrderSortBy(query.Get("sortBy"))}
	if params.SortBy != "" && params.SortBy != domain.OrderSortByDescDate && params.SortBy != domain.OrderSortByAscDate {
		return domain.OrdersParams{}, errors.New("некорректная сортировка заказов")
	}
	if rawLimit := query.Get("limit"); rawLimit != "" {
		limit, err := strconv.ParseUint(rawLimit, 10, 32)
		if err != nil || limit == 0 {
			return domain.OrdersParams{}, errors.New("некорректный лимит заказов")
		}
		params.Limit = uint(limit)
	}
	if rawCursor := query.Get("cursor"); rawCursor != "" {
		decoded, err := decodeCursor(rawCursor)
		if err != nil {
			return domain.OrdersParams{}, err
		}
		params.Cursor = &decoded
	}
	return params, nil
}

func encodeCursor(cursor domain.Cursor) string {
	b, _ := json.Marshal(orderCursorJSON{CreatedAt: cursor.CreatedAt.UTC(), ID: cursor.ID})
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(value string) (domain.Cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return domain.Cursor{}, errors.New("некорректный курсор заказов")
	}
	var decoded orderCursorJSON
	if err := json.Unmarshal(raw, &decoded); err != nil || decoded.ID <= 0 || decoded.CreatedAt.IsZero() {
		return domain.Cursor{}, errors.New("некорректный курсор заказов")
	}
	return domain.Cursor{CreatedAt: decoded.CreatedAt, ID: decoded.ID}, nil
}
