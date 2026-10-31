package domain

type ProductCategory struct {
	ID         int64
	Category   string
	PictureURL *string
}

type CategoriesSearchParams struct {
	Limit  uint
	SortBy SortByType
}
