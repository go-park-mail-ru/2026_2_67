package delivery

import (
	"context"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/sellers/domain"
	"vibe_market/backend/internal/sellers/usecase"

	"go.uber.org/zap"
)

type SellersHandler struct {
	sellerRepo domain.SellerRepo
	usecase    usecase.SellerUsecase
	baseLogger *zap.Logger
}

func NewSellersHandler(sellerRepo domain.SellerRepo, usecase usecase.SellerUsecase, baseLogger *zap.Logger) *SellersHandler {
	return &SellersHandler{sellerRepo: sellerRepo, usecase: usecase, baseLogger: baseLogger}
}

func (h *SellersHandler) logger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(common.LoggerContextKey).(*zap.Logger)
	if !ok {
		return h.baseLogger
	}
	return logger
}
