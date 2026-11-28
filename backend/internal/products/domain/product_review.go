package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const OrderStatusIDCanPostReview = 4

type ProductReviewSortBy string

const (
	ReviewSortByDescDate   ProductReviewSortBy = "desc-date"
	ReviewSortByAscDate    ProductReviewSortBy = "asc-date"
	ReviewSortByDescRating ProductReviewSortBy = "desc-rating"
	ReviewSortByAscRating  ProductReviewSortBy = "asc-rating"
)

var (
	ErrProductReviewSortByInvalid = fmt.Errorf("sortBy должен быть одним из: [%s, %s, %s, %s]",
		ReviewSortByDescDate, ReviewSortByAscDate, ReviewSortByDescRating, ReviewSortByAscRating)
	ErrProductReviewCursorInvalid = errors.New("некорректный курсор отзывов")
	ErrProductReviewLimitInvalid  = errors.New("limit должен быть положительным числом")
	ErrProductReviewNotFound      = errors.New("отзыв на продукт не найден")
)

type ProductReview struct {
	BuyerID   int64
	ProductID int64
	Review    *string
	Rating    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ProductReviewListItem struct {
	BuyerID         int64
	BuyerName       string
	BuyerSurname    *string
	BuyerPatronymic *string
	BuyerAvatarURL  *string
	Review          *string
	Rating          int
	CreatedAt       time.Time
}

type ProductReviewCursor struct {
	CreatedAt time.Time
	BuyerID   int64
	Rating    int
}

type ProductReviewsParams struct {
	ProductID int64
	SortBy    ProductReviewSortBy
	Limit     int
	Cursor    *ProductReviewCursor
}

type ProductReviewsPage struct {
	Reviews    []ProductReviewListItem
	NextCursor *ProductReviewCursor
}

type ProductReviewRepo interface {
	Insert(pr ProductReview) error
	SelectByProductID(ctx context.Context, params ProductReviewsParams) (ProductReviewsPage, error)
	SelectByProductIDAndBuyerID(ctx context.Context, productID, buyerID int64) (ProductReview, error)
	Update(ctx context.Context, pr ProductReview) (ProductReview, error)
}
