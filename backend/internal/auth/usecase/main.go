package usecase

import (
	"vibe_market/backend/internal/auth/domain"
	buyerDomain "vibe_market/backend/internal/buyers/domain"
)

type AuthUsecase struct {
	jwtMeta   domain.JwtAuthMeta
	userRepo  domain.UserRepo
	tokenRepo domain.RefreshTokenRepo
	buyerRepo buyerDomain.BuyerRepo
}

func NewAuthUsecase(jwtMeta domain.JwtAuthMeta, userRepo domain.UserRepo,
	tokenRepo domain.RefreshTokenRepo, buyerRepo buyerDomain.BuyerRepo) *AuthUsecase {
	return &AuthUsecase{jwtMeta, userRepo, tokenRepo, buyerRepo}
}
