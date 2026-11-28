package domain

import "fmt"

type Category struct {
	ID         int64
	Category   string
	PictureURL *string
}

type SortByType string

const (
	SortByNew     SortByType = "new"
	SortByPopular SortByType = "popular"
)

var (
	ErrSortByInvalid = fmt.Errorf("sortBy может быть одним из: [%s, %s]", SortByNew, SortByPopular)
)

type CategoriesSearchParams struct {
	Limit     uint
	SortBy    SortByType
	Seller    []string
	Query     string
	PriceMin  float64
	PriceMax  *float64
	RatingMin float64
	RatingMax *float64
}

type CategoryRepo interface {
	Search(params CategoriesSearchParams) ([]Category, error)
	SearchByQuery(query string, limit uint) ([]Category, error)
}
