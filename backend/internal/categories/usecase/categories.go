package usecase

import (
	"errors"
	"vibe_market/backend/internal/categories/domain"
)

type CategoriesOutDTO struct {
	Categories []domain.Category
}

func Categories(dto domain.CategoriesSearchParams, categoryRepo domain.CategoryRepo) (CategoriesOutDTO, error) {
	categories, err := categoryRepo.Search(dto)
	if errors.Is(err, domain.ErrInvalidSortBy) {
		return CategoriesOutDTO{}, err
	}
	return CategoriesOutDTO{Categories: categories}, err
}
