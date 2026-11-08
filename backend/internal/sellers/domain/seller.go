package domain

import "time"

type Seller struct {
	UserID      int64
	Name        string
	Description *string
	AvatarURL   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SellerSearchParams struct {
	Limit    uint
	Category string
	Query    string
	Offset   uint
	PriceMin float64
	PriceMax *float64
}

type SellerRepo interface {
	Search(params SellerSearchParams) ([]Seller, error)
}
