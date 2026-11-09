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
	ErrOrderBasketEmpty       = errors.New("корзина пуста")
	ErrOrderStockInsufficient = errors.New("товары закончились или их недостаточно")
	ErrPromocodeInvalid       = errors.New("промокод не найден или истек")
	ErrPickupPointInvalid     = errors.New("пункт выдачи не найден")
)

type OrderDetails struct {
	ID            int64
	Status        OrderStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	PickupPoint   OrderPickupPoint
	Promocode     *OrderPromocode
	Products      []OrderProductDetails
	Notifications []OrderNotification
}

type OrderStatus struct {
	ID   int64
	Name string
}

type OrderPickupPoint struct {
	ID        int64
	Longitude float64
	Latitude  float64
	StartTime string
	EndTime   string
}

type OrderPromocode struct {
	ID                  int64
	Code                string
	Discount            int
	AvailableCategoryID *int64
	StartDatetime       time.Time
	EndDatetime         time.Time
}

type OrderProductDetails struct {
	ID          int64
	ProductID   int64
	Name        string
	Description *string
	Category    OrderCategory
	Seller      OrderSeller
	PictureURLs []string
	Count       int
	Price       int64
}

type OrderCategory struct {
	ID   int64
	Name string
}

type OrderSeller struct {
	ID   int64
	Name string
}

type OrderNotification struct {
	ID        int64
	Status    OrderStatus
	IsRead    bool
	CreatedAt time.Time
}

type OrderRepo interface {
	ListByBuyer(ctx context.Context, buyerID int64) ([]OrderDetails, error)
	CreateOrderTx(ctx context.Context, buyerID, pickupPointID int64, promoCode string) (int64, []ProductStockError, error)
}
