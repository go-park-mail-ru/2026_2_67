package usecase

import (
	"vibe_market/backend/internal/auth/domain"
	buyerDomain "vibe_market/backend/internal/buyers/domain"
)

type AuthUsecase interface {
	Login(dto LoginInDTO) (LoginOutDTO, error)
	Logout(accessToken domain.AccessTokenPayload, deviceName string, tokenRepo domain.RefreshTokenRepo) error
	Refresh(dto RefreshInDTO) (RefreshOutDTO, error)
	Register(dto RegisterInDTO) (RegisterOutDTO, error)
}

type AuthUsecaseImpl struct {
	jwtMeta   domain.JwtAuthMeta
	userRepo  domain.UserRepo
	tokenRepo domain.RefreshTokenRepo
	buyerRepo buyerDomain.BuyerRepo
}

func NewAuthUsecase(jwtMeta domain.JwtAuthMeta, userRepo domain.UserRepo,
	tokenRepo domain.RefreshTokenRepo, buyerRepo buyerDomain.BuyerRepo) *AuthUsecaseImpl {
	return &AuthUsecaseImpl{jwtMeta, userRepo, tokenRepo, buyerRepo}
}
