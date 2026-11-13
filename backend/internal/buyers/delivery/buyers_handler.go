package delivery

import (
	"context"
	"net/http"
	authDomain "vibe_market/backend/internal/auth/domain"
	authMiddleware "vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

type BuyersHandler struct {
	buyerRepo  domain.BuyerRepo
	basketRepo domain.BasketRepo
	userRepo   authDomain.UserRepo
	jwtSecret  []byte
	avatarRepo domain.BuyerAvatarRepo
	baseLogger *zap.Logger
}

func NewBuyersHandler(
	buyerRepo domain.BuyerRepo,
	basketRepo domain.BasketRepo,
	userRepo authDomain.UserRepo,
	jwtSecret []byte,
	baseLogger *zap.Logger,
	avatarRepo domain.BuyerAvatarRepo,
) *BuyersHandler {
	return &BuyersHandler{
		buyerRepo:  buyerRepo,
		basketRepo: basketRepo,
		userRepo:   userRepo,
		jwtSecret:  jwtSecret,
		baseLogger: baseLogger,
		avatarRepo: avatarRepo,
	}
}

func (h *BuyersHandler) logger(ctx context.Context) *zap.Logger {
	if val := ctx.Value(common.LoggerContextKey); val != nil {
		if l, ok := val.(*zap.Logger); ok && l != nil {
			return l
		}
	}
	return h.baseLogger
}

func (h *BuyersHandler) authenticate(r *http.Request) (*authDomain.AccessTokenPayload, int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, http.StatusBadRequest, authDomain.ErrUserIDNotFound
	}

	payload, err := authMiddleware.ParseAccessTokenPayload(r, h.jwtSecret)
	if err != nil || payload == nil {
		return nil, http.StatusForbidden, err
	}

	return payload, http.StatusOK, nil
}
