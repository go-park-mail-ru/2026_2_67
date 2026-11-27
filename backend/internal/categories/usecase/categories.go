package usecase

import (
	"errors"
	"vibe_market/backend/internal/categories/domain"
)

type CategoriesOutDTO struct {
	Categories []domain.Category
}

func (u *CategoryUsecaseImpl) Categories(dto domain.CategoriesSearchParams) (CategoriesOutDTO, error) {
	if dto.SortBy == "" {
		dto.SortBy = domain.SortByNew
	}
	dto.PriceMin *= 100
	if dto.PriceMax != nil {
		*dto.PriceMax *= 100
	}
	categories, err := u.categoryRepo.Search(dto)
	if errors.Is(err, domain.ErrSortByInvalid) {
		return CategoriesOutDTO{}, err
	}
	return CategoriesOutDTO{Categories: categories}, err
}
