// Package controller для аутентификации/авторизации пользователей
package controller

import (
	"vibe_market/backend/internal/auth/domain"
	buyerDomain "vibe_market/backend/internal/buyers/domain"
)

type AuthHandler struct {
	jwtMeta   domain.JwtAuthMeta
	userRepo  domain.UserRepo
	tokenRepo domain.RefreshTokenRepo
	buyerRepo buyerDomain.BuyerRepo
}

func NewAuthHandler(jwtMeta domain.JwtAuthMeta, userRepo domain.UserRepo,
	tokenRepo domain.RefreshTokenRepo, buyerRepo buyerDomain.BuyerRepo) *AuthHandler {
	return &AuthHandler{jwtMeta, userRepo, tokenRepo, buyerRepo}
}
