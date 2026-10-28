package domain

import (
	"context"
	"errors"
	"time"
)

type Order struct {
	ID            int64
	BuyerID       int64
	OrderStatusID int64
	PromocodeID   *int64
	PickupPointID int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type OrderProduct struct {
	ID        int64
	OrderID   int64
	ProductID int64
	Count     int
	Price     int
	CreatedAt time.Time
}

type ProductStockError struct {
	ProductID  int64
	Name       string
	Seller     string
	ErrMessage string
}

var (
	ErrOrderBasketEmpty         = errors.New("корзина пуста")
	ErrOrderStockInsufficient   = errors.New("товары закончились или их недостаточно")
	ErrPromocodeInvalid         = errors.New("промокод не найден или истек")
	ErrInvalidCoordinates       = errors.New("неверные координаты")
	ErrNoPickupPointsAvailable  = errors.New("нет доступных пунктов выдачи")
)

type OrderRepo interface {
	CreateOrderTx(ctx context.Context, buyerID int64, lat, lng float64, promoCode string) (int64, []ProductStockError, error)
}
