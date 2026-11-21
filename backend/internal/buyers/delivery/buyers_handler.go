package delivery

import (
	"context"
	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/buyers/usecase"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

type BuyersHandler struct {
	buyerRepo  domain.BuyerRepo
	basketRepo domain.BasketRepo
	userRepo   authDomain.UserRepo
	avatarRepo domain.BuyerAvatarRepo
	usecase    usecase.BuyerUsecase
	baseLogger *zap.Logger
}

func NewBuyersHandler(
	buyerRepo domain.BuyerRepo,
	basketRepo domain.BasketRepo,
	userRepo authDomain.UserRepo,
	baseLogger *zap.Logger,
	avatarRepo domain.BuyerAvatarRepo,
	usecase usecase.BuyerUsecase,
) *BuyersHandler {
	return &BuyersHandler{
		buyerRepo:  buyerRepo,
		basketRepo: basketRepo,
		userRepo:   userRepo,
		baseLogger: baseLogger,
		avatarRepo: avatarRepo,
		usecase:    usecase,
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
