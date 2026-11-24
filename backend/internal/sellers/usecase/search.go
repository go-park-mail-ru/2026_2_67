package usecase

import "vibe_market/backend/internal/sellers/domain"

type SearchOutDTO struct {
	Sellers []domain.Seller
}

func (u *SellerUsecaseImpl) Search(query string, limit uint) (SearchOutDTO, error) {
	sellers, err := u.sellerRepo.SearchByQuery(query, limit)
	if err != nil {
		return SearchOutDTO{}, err
	}
	return SearchOutDTO{Sellers: sellers}, nil
}
