package usecase

import "vibe_market/backend/internal/products/domain"

type SearchOutDTO struct {
	Products []domain.ProductSearchResult
}

func (u *ProductUsecaseImpl) Search(query string, limit uint) (SearchOutDTO, error) {
	if limit == 0 {
		limit = productsDefaultLimit
	}
	if limit > productsMaxLimit {
		limit = productsMaxLimit
	}
	products, err := u.productRepo.SearchByQuery(query, limit)
	if err != nil {
		return SearchOutDTO{}, err
	}
	return SearchOutDTO{Products: products}, nil
}
