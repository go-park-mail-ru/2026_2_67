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
	Price          int
	Discount       uint
	Name           string
	Description    *string
	AvailableCount uint
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Cursor struct {
	CreatedAt time.Time
	BuyCount  uint
	Price     int
	ID        int64
	Rating    float64
}

type SortByType string

const (
	SortByBuyCount   SortByType = "buy-count"
	SortByAscPrice   SortByType = "asc-price"
	SortByDescPrice  SortByType = "desc-price"
	SortByHighRating SortByType = "high-rating"
	SortByNew        SortByType = "new"
)

var (
	ErrPriceRangeInvalid = errors.New("цены должны быть конечными и неотрицательными, priceMin должен быть <= priceMax")
	ErrCursorInvalid     = errors.New("курсор должен содержать дату создания и положительный id")
	ErrSortByInvalid     = fmt.Errorf("sortBy должен быть одним из: [%s, %s, %s, %s, %s]",
		SortByBuyCount, SortByAscPrice, SortByDescPrice, SortByHighRating, SortByNew)
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
	BuyCounts    []uint
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

var (
	ErrProductReviewAlreadyExists = errors.New("отзыв на продукт уже создан")
	ErrProductReviewRatingInvalid = errors.New("отзыв на продукт имеет неверный рейтинг")
)
