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
	ErrInvalidSortBy = fmt.Errorf("sortBy может быть одним из: [%s, %s]", SortByNew, SortByPopular)
)

type CategoriesSearchParams struct {
	Limit  uint
	SortBy SortByType
}

type CategoryRepo interface {
	Search(params CategoriesSearchParams) ([]Category, error)
}
