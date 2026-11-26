package usecase

import "vibe_market/backend/internal/sellers/domain"

type SellersOutDTO struct {
	Sellers []domain.Seller
}

const defaultSellersLimit = 10

func (u *SellerUsecaseImpl) Sellers(dto domain.SellerSearchParams) (SellersOutDTO, error) {
	if dto.Limit == 0 {
		dto.Limit = defaultSellersLimit
	}
	var (
		answer SellersOutDTO
		err    error
	)
	answer.Sellers, err = u.sellerRepo.Search(dto)
	return answer, err
}
