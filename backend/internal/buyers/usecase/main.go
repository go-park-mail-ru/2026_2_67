package usecase

import (
	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/buyers/domain"
)

type BuyerUsecase interface {
	Buyer(userID int64) (BuyerOutDTO, error)
	PatchBuyer(dto PatchBuyerInDTO) (BuyerOutDTO, error)
	PatchBuyerAvatar(dto PatchBuyerAvatarInDTO) error
}

type BuyerUsecaseImpl struct {
	buyerRepo  domain.BuyerRepo
	avatarRepo domain.BuyerAvatarRepo
	userRepo   authDomain.UserRepo
}

func NewBuyerUsecase(buyerRepo domain.BuyerRepo, avatarRepo domain.BuyerAvatarRepo, userRepo authDomain.UserRepo) *BuyerUsecaseImpl {
	return &BuyerUsecaseImpl{buyerRepo, avatarRepo, userRepo}
}
