package usecase

import "vibe_market/backend/internal/sellers/domain"

type SellerUsecase interface {
	Sellers(dto domain.SellerSearchParams) (SellersOutDTO, error)
}

type SellerUsecaseImpl struct {
	sellerRepo domain.SellerRepo
}

func NewSellerUsecase(sellerRepo domain.SellerRepo) *SellerUsecaseImpl {
	return &SellerUsecaseImpl{sellerRepo}
}
