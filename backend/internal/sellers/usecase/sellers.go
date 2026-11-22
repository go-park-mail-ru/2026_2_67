package usecase

import "vibe_market/backend/internal/sellers/domain"

type SellersOutDTO struct {
	Sellers []domain.Seller
}

func (u *SellerUsecaseImpl) Sellers(dto domain.SellerSearchParams) (SellersOutDTO, error) {
	var (
		answer SellersOutDTO
		err    error
	)
	answer.Sellers, err = u.sellerRepo.Search(dto)
	return answer, err
}
