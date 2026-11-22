package usecase

import (
	"errors"
	"vibe_market/backend/internal/categories/domain"
)

type CategoriesOutDTO struct {
	Categories []domain.Category
}

func (u *CategoryUsecaseImpl) Categories(dto domain.CategoriesSearchParams) (CategoriesOutDTO, error) {
	categories, err := u.categoryRepo.Search(dto)
	if errors.Is(err, domain.ErrSortByInvalid) {
		return CategoriesOutDTO{}, err
	}
	return CategoriesOutDTO{Categories: categories}, err
}
