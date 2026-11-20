package domain

import (
	"errors"
	"fmt"
	"time"
	"vibe_market/backend/internal/sellers/domain"
)

type Product struct {
	ID             int64
	CategoryID     int64
	SellerID       int64
	Price          float64
	Discount       uint
	Name           string
	Description    *string
	AvailableCount uint
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

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

var (
	ErrPriceRangeInvalid = errors.New("цены должны быть конечными и неотрицательными, priceMin должен быть <= priceMax")
	ErrCursorInvalid     = errors.New("курсор должен содержать дату создания и положительный id")
	ErrSortByInvalid     = fmt.Errorf("sortBy должен быть одним из: [%s, %s]", SortByNew, SortByPopular)
	ErrProductIDNotFound = errors.New("продукт с таким id не найден")
)

type ProductsSearchParams struct {
	Category string
	Seller   []string
	Query    string
	Cursor   *Cursor
	Limit    uint
	PriceMin float64
	PriceMax *float64
	SortBy   SortByType
}

type ProductsPage struct {
	Products     []Product
	Sellers      []domain.Seller
	PictureURLs  [][]string
	Ratings      []float64
	ReviewCounts []uint
	NextCursor   *Cursor
}

type ProductByID struct {
	Product
	Category        string
	SellerName      string
	SellerAvatarURL *string
	PictureURLS     []string
	Rating          float64
	MarksCount      uint
	ReviewsCount    uint
}

type ProductRepo interface {
	Search(params ProductsSearchParams) (ProductsPage, error)
	SelectByID(id int64) (ProductByID, error)
}

type ProductOutDTO struct {
	ProductByID
	PriceResult float64
}

type ProductsOutDTO struct {
	Products     []Product
	Sellers      []domain.Seller
	PricesResult []float64
	PictureURLs  [][]string
	Ratings      []float64
	ReviewCounts []uint
	NextCursor   *Cursor
}

var (
	ErrProductReviewAlreadyExists = errors.New("отзыв на продукт уже создан")
	ErrProductReviewRatingInvalid = errors.New("отзыв на продукт имеет неверный рейтинг")
)

type ProductUsecase interface {
	Product(id int64) (ProductOutDTO, error)
	Products(dto ProductsSearchParams) (ProductsOutDTO, error)
}
