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

type ProductBasketError struct {
	ProductID  int64
	Name       string
	Seller     string
	ErrMessage string
}

var (
	ErrOrderBasketEmpty       = errors.New("корзина пуста")
	ErrOrdersNotFound         = errors.New("заказы не найдены")
	ErrOrderStockInsufficient = errors.New("товары закончились или их недостаточно")
	ErrPromocodeInvalid       = errors.New("промокод не найден или истек")
	ErrPickupPointInvalid     = errors.New("пункт выдачи не найден")
)

type OrderProduct struct {
	OrderID   int64
	ProductID int64
	Count     uint
	Price     int
	CreatedAt time.Time
}

type OrderRepo interface {
	Orders(ctx context.Context, buyerID int64, params OrdersParams) (OrderPage, error)
	CreateOrderTx(ctx context.Context, buyerID, pickupPointID int64, promoCode string) (int64, []ProductBasketError, error)
	SelectByProductIDAndBuyerID(productID int64, buyerID int64) ([]Order, error)
}
