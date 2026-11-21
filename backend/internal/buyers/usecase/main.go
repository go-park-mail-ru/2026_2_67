package usecase

import (
	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/buyers/domain"
)

type BuyerUsecase struct {
	buyerRepo  domain.BuyerRepo
	avatarRepo domain.BuyerAvatarRepo
	userRepo   authDomain.UserRepo
}

func NewBuyerUsecase(buyerRepo domain.BuyerRepo, avatarRepo domain.BuyerAvatarRepo, userRepo authDomain.UserRepo) *BuyerUsecase {
	return &BuyerUsecase{buyerRepo, avatarRepo, userRepo}
}
