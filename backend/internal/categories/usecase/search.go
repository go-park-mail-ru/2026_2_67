package usecase

import "vibe_market/backend/internal/categories/domain"

type SearchOutDTO struct {
	Categories []domain.Category
}

func (u *CategoryUsecaseImpl) Search(query string, limit uint) (SearchOutDTO, error) {
	categories, err := u.categoryRepo.SearchByQuery(query, limit)
	if err != nil {
		return SearchOutDTO{}, err
	}
	return SearchOutDTO{Categories: categories}, nil
}
