package usecase

import "vibe_market/backend/internal/sellers/domain"

type SellersOutDTO struct {
	Sellers []domain.Seller
}

func Sellers(dto domain.SellerSearchParams, repo domain.SellerRepo) (SellersOutDTO, error) {
	var (
		answer SellersOutDTO
		err    error
	)
	answer.Sellers, err = repo.Search(dto)
	return answer, err
}
