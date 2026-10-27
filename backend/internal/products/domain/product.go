package domain

import (
	"errors"
	"time"
)

type Product struct {
	ID             int64
	CategoryID     int64
	SellerID       int64
	Price          float64
	Discount       int
	Name           string
	Description    *string
	AvailableCount int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

var (
	ErrInvalidPriceRange = errors.New("цены должны быть конечными и неотрицательными, priceMin должен быть <= priceMax")
	ErrInvalidCursor     = errors.New("курсор должен содержать дату создания и положительный id")
	ErrInvalidSortBy     = errors.New("sortBy должен быть new или popular")
)

type Cursor struct {
	CreatedAt time.Time
	ID        int64
	Rating    float64 // Используется для курсора сортировки popular.
}

type SortByType string

const (
	SortByPopular SortByType = "popular"
	SortByNew     SortByType = "new"
)

type ProductsSearchParams struct {
	Category string
	Seller   string
	Query    string
	Cursor   Cursor
	Limit    uint
	PriceMin float64
	PriceMax *float64 // nil означает отсутствие верхней границы.
	SortBy   SortByType
}

type ProductsPage struct {
	Products     []Product
	PictureURLs  [][]string
	Ratings      []float32
	ReviewCounts []uint
	NextCursor   *Cursor // nil, если следующей страницы нет.
}

type ProductRepo interface {
	Search(params ProductsSearchParams) (ProductsPage, error)
	SearchCategories(params CategoriesSearchParams) ([]ProductCategory, error)
}
