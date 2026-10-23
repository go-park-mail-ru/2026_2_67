// Package controller для аутентификации/авторизации пользователей
package controller

import (
	"vibe_market/backend/internal/auth/domain"
)

type AuthHandler struct {
	jwtMeta   domain.JwtAuthMeta
	userRepo  domain.UserRepo
	tokenRepo domain.RefreshTokenRepo
}

func NewAuthHandler(jwtMeta domain.JwtAuthMeta, userRepo domain.UserRepo,
	tokenRepo domain.RefreshTokenRepo) *AuthHandler {
	return &AuthHandler{jwtMeta, userRepo, tokenRepo}
}
