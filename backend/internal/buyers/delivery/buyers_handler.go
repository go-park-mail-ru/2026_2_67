package delivery

import (
	"context"
	"vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

type BuyersHandler struct {
	buyerRepo  domain.BuyerRepo
	baseLogger *zap.Logger
}

func NewBuyersHandler(buyerRepo domain.BuyerRepo, baseLogger *zap.Logger) *BuyersHandler {
	return &BuyersHandler{buyerRepo, baseLogger}
}

func (h *BuyersHandler) logger(ctx context.Context) *zap.Logger {
	advanced := ctx.Value(common.LoggerContextKey).(*zap.Logger)
	if advanced != nil {
		return advanced
	}
	return h.baseLogger
}
