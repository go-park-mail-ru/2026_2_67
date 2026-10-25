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
	advanced := ctx.Value(common.LoggerContextKey).(*zap.Logger)
	if advanced == nil {
		return h.baseLogger
	}
	return advanced
}
