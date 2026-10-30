package delivery

import (
	"context"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/sellers/domain"

	"go.uber.org/zap"
)

type SellersHandler struct {
	sellerRepo domain.SellerRepo
	baseLogger *zap.Logger
}

func NewSellersHandler(sellerRepo domain.SellerRepo, baseLogger *zap.Logger) *SellersHandler {
	return &SellersHandler{sellerRepo: sellerRepo, baseLogger: baseLogger}
}

func (h *SellersHandler) logger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(common.LoggerContextKey).(*zap.Logger)
	if !ok {
		return h.baseLogger
	}
	return logger
}
