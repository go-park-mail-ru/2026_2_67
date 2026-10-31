package usecase

import (
	"errors"
	"vibe_market/backend/internal/products/domain"
)

type ProductCategoriesOutDTO struct {
	Categories []domain.ProductCategory
}

func ProductCategories(dto domain.CategoriesSearchParams, categoryRepo domain.ProductRepo) (ProductCategoriesOutDTO, error) {
	categories, err := categoryRepo.SearchCategories(dto)
	if errors.Is(err, domain.ErrInvalidSortBy) {
		return ProductCategoriesOutDTO{}, err
	}
	return ProductCategoriesOutDTO{Categories: categories}, err
}
