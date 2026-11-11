package domain

import (
	"errors"
	"time"
)

type BasketItem struct {
	ProductID      int64
	CategoryName   string
	Name           string
	PictureURL     string
	Price          float64
	Discount       int
	PriceResult    float64
	Count          int
	AvailableCount int
	SellerName     string
	CategoryID     int64
}

type BasketSummary struct {
	Items           []BasketItem
	BasketSum       float64
	BasketSumResult float64
}

type PickupPoint struct {
	UserID    int64
	Longitude float64
	Latitude  float64
	StartTime string
	EndTime   string
}

type Promocode struct {
	ID                  int64
	Code                string
	AvailableCategoryID *int64
	Discount            int
	StartDatetime       time.Time
	EndDatetime         time.Time
}

type ProductStockError struct {
	ProductID  int64
	Name       string
	Seller     string
	ErrMessage string
}

var (
	ErrBasketEmpty              = errors.New("корзина пуста")
	ErrCountInvalid             = errors.New("неверное количество товара")
	ErrCountExceeded            = errors.New("запрашиваемое количество превышает доступный остаток")
	ErrProductNotFoundInBasket  = errors.New("товар не найден в корзине")
	ErrProductNotFound          = errors.New("товар не найден")
	ErrBuyerNotFound            = errors.New("покупатель не найден")
	ErrPromocodeNotFoundOrExpired = errors.New("промокод не найден или истек")
	ErrInvalidCoordinates       = errors.New("неверные координаты доставки")
	ErrNoPickupPoints           = errors.New("нет доступных пунктов выдачи")
)

type BasketRepo interface {
	GetBasket(buyerID int64) (BasketSummary, error)
	AddProduct(buyerID int64, productID int64, count int) error
	UpdateProductCount(buyerID int64, productID int64, count int) error
	GetNearestPickupPoint(lat, lng float64) (PickupPoint, error)
	GetPromocode(code string) (*Promocode, error)
}
