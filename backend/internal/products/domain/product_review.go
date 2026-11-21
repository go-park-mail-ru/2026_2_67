package domain

import "time"

const OrderStatusIDCanPostReview = 4

type ProductReview struct {
	BuyerID   int64
	ProductID int64
	Review    *string
	Rating    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ProductReviewRepo interface {
	Insert(pr ProductReview) error
}
