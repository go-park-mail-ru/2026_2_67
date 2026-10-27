package domain

import "errors"

type ProductCategory struct {
	ID         int64
	Category   string
	PictureURL *string
}

var (
	ErrInvalidLimit = errors.New("limit должен быть >= 0")
)

type CategoriesSearchParams struct {
	Limit  int
	SortBy SortByType
}
