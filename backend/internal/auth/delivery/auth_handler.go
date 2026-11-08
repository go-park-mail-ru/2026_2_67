// Package controller для аутентификации/авторизации пользователей
package delivery

import (
	"context"
	"vibe_market/backend/internal/auth/domain"
	buyerDomain "vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

type AuthHandler struct {
	jwtMeta    domain.JwtAuthMeta
	userRepo   domain.UserRepo
	tokenRepo  domain.RefreshTokenRepo
	buyerRepo  buyerDomain.BuyerRepo
	baseLogger *zap.Logger
}

func NewAuthHandler(jwtMeta domain.JwtAuthMeta, userRepo domain.UserRepo,
	tokenRepo domain.RefreshTokenRepo, buyerRepo buyerDomain.BuyerRepo, baseLogger *zap.Logger) *AuthHandler {
	return &AuthHandler{jwtMeta, userRepo, tokenRepo, buyerRepo, baseLogger}
}

func (h *AuthHandler) logger(ctx context.Context) *zap.Logger {
	if val := ctx.Value(common.LoggerContextKey); val != nil {
		if advanced, ok := val.(*zap.Logger); ok && advanced != nil {
			return advanced
		}
	}
	return h.baseLogger
}
