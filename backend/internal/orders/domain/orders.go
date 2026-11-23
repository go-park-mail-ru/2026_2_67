package domain

import (
	"context"
	"time"
	pickupPointDomain "vibe_market/backend/internal/pickup_points/domain"
	sellerDomain "vibe_market/backend/internal/sellers/domain"
)

type OrderView struct {
	ID          int64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PickupPoint pickupPointDomain.PickupPoint
	Promocode   *string
	Products    []OrderProductView
	TotalCount  uint
	TotalPrice  int
}

type OrderProductView struct {
	OrderProduct
	Name        string
	Description string
	Category    string
	Seller      sellerDomain.Seller
	PictureURLS []string
	TotalPrice  int
}

type OrderSortBy string

const (
	OrderSortByDescDate OrderSortBy = "desc-date"
	OrderSortByAscDate  OrderSortBy = "asc-date"
)

type Cursor struct {
	CreatedAt time.Time
	ID        int64
}

type OrdersParams struct {
	Status string
	SortBy OrderSortBy
	Limit  uint
	Cursor *Cursor
}

type OrderPage struct {
	Orders     []OrderView
	NextCursor *Cursor
}

type OrderListRepo interface {
	OrdersPage(ctx context.Context, buyerID int64, params OrdersParams) (OrderPage, error)
}
