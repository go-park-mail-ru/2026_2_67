package usecase

import "vibe_market/backend/internal/categories/domain"

type CategoryUsecase interface {
	Categories(dto domain.CategoriesSearchParams) (CategoriesOutDTO, error)
}

type CategoryUsecaseImpl struct {
	categoryRepo domain.CategoryRepo
}

func NewCategoryUsecase(categoryRepo domain.CategoryRepo) *CategoryUsecaseImpl {
	return &CategoryUsecaseImpl{categoryRepo}
}
