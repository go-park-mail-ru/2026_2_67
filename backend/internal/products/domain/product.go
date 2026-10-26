package domain

import (
	"errors"
	"fmt"
	"time"
)

type Product struct {
	ID             int64
	CategoryID     int64
	SellerID       int64
	Price          int
	Name           string
	Description    *string
	AvailableCount int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

const DefaultLimit = 20
const MaxLimit = 30

var (
	ErrLimitOverflow     = fmt.Errorf("limit должен быть <= %d", MaxLimit)
	ErrInvalidPriceRange = errors.New("цены должны быть конечными и неотрицательными, priceMin должен быть <= priceMax")
	ErrInvalidCursor     = errors.New("курсор должен содержать дату создания и положительный id")
)

type Cursor struct {
	CreatedAt time.Time
	ID        int64
}

type SearchParams struct {
	Category string
	Seller   string
	Query    string
	Cursor   Cursor
	Limit    uint
	PriceMin float32
	PriceMax *float32 // nil означает отсутствие верхней границы.
}

// Данные в PictureURLs, Ratings и ReviewCounts относятся к Products с тем же индексом.
type ProductsPage struct {
	Products     []Product
	PictureURLs  [][]string
	Ratings      []float32
	ReviewCounts []uint
	NextCursor   *Cursor // nil, если следующей страницы нет.
}

type ProductRepo interface {
	Search(params SearchParams) (ProductsPage, error)
}
