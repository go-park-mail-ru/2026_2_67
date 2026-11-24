package usecase

import "vibe_market/backend/internal/categories/domain"

type CategoryUsecase interface {
	Categories(dto domain.CategoriesSearchParams) (CategoriesOutDTO, error)
	Search(query string, limit uint) (SearchOutDTO, error)
}

type CategoryUsecaseImpl struct {
	categoryRepo domain.CategoryRepo
}

func NewCategoryUsecase(categoryRepo domain.CategoryRepo) *CategoryUsecaseImpl {
	return &CategoryUsecaseImpl{categoryRepo}
}
