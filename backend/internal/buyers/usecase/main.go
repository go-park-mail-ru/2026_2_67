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

type BasketUsecase interface {
	Basket(buyerID int64) (BasketOutDTO, error)
	IsValid(buyerID int64) (IsValidOutDTO, error)
	AddProduct(buyerID int64, productID int64, count int) error
	UpdateProductCount(buyerID int64, productID int64, count int) error
	PreviewBasket(buyerID int64, dto PreviewBasketInDTO) (PreviewBasketOutDTO, PreviewBasketValidationErrors, []domain.ProductStockError, error)
}

type BuyerUsecaseImpl struct {
	buyerRepo  domain.BuyerRepo
	avatarRepo domain.BuyerAvatarRepo
	userRepo   authDomain.UserRepo
}

func NewBuyerUsecase(buyerRepo domain.BuyerRepo, avatarRepo domain.BuyerAvatarRepo, userRepo authDomain.UserRepo) *BuyerUsecaseImpl {
	return &BuyerUsecaseImpl{buyerRepo, avatarRepo, userRepo}
}

type BasketUsecaseImpl struct {
	basketRepo domain.BasketRepo
}

func NewBasketUsecase(basketRepo domain.BasketRepo) *BasketUsecaseImpl {
	return &BasketUsecaseImpl{basketRepo: basketRepo}
}
