package delivery

import (
	"context"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/products/domain"

	"go.uber.org/zap"
)

type ProductsHandler struct {
	productRepo domain.ProductRepo
	baseLogger  *zap.Logger
}

func NewProductsHandler(productRepo domain.ProductRepo, logger *zap.Logger) *ProductsHandler {
	return &ProductsHandler{productRepo: productRepo, baseLogger: logger}
}

func (h *ProductsHandler) logger(ctx context.Context) *zap.Logger {
	logger := ctx.Value(common.LoggerContextKey).(*zap.Logger)
	if logger != nil {
		return logger
	}
	return h.baseLogger
}
