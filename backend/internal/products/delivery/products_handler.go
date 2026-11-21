package delivery

import (
	"context"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/products/domain"
	"vibe_market/backend/internal/products/usecase"

	"go.uber.org/zap"
)

type ProductsHandler struct {
	productRepo          domain.ProductRepo
	productUsecase       usecase.ProductUsecase
	productReviewUsecase usecase.ProductReviewUsecase
	baseLogger           *zap.Logger
	jwtSecret            []byte
}

func NewProductsHandler(productRepo domain.ProductRepo, productUsecase usecase.ProductUsecase,
	productReviewUsecase usecase.ProductReviewUsecase, logger *zap.Logger, jwtSecret []byte) *ProductsHandler {
	return &ProductsHandler{productRepo: productRepo, productUsecase: productUsecase,
		productReviewUsecase: productReviewUsecase, baseLogger: logger, jwtSecret: jwtSecret}
}

func (h *ProductsHandler) logger(ctx context.Context) *zap.Logger {
	logger := ctx.Value(common.LoggerContextKey).(*zap.Logger)
	if logger != nil {
		return logger
	}
	return h.baseLogger
}
