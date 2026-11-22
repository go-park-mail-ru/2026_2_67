package delivery

import (
	"context"
	"vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/buyers/usecase"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

type BuyersHandler struct {
	avatarRepo    domain.BuyerAvatarRepo
	buyerUsecase  usecase.BuyerUsecase
	basketUsecase usecase.BasketUsecase
	baseLogger    *zap.Logger
}

func NewBuyersHandler(
	baseLogger *zap.Logger,
	avatarRepo domain.BuyerAvatarRepo,
	buyerUsecase usecase.BuyerUsecase,
	basketUsecase usecase.BasketUsecase,
) *BuyersHandler {
	return &BuyersHandler{
		baseLogger:    baseLogger,
		avatarRepo:    avatarRepo,
		buyerUsecase:  buyerUsecase,
		basketUsecase: basketUsecase,
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
