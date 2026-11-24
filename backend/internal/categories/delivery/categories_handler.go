package delivery

import (
	"context"
	"vibe_market/backend/internal/categories/domain"
	"vibe_market/backend/internal/categories/usecase"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

type CategoriesHandler struct {
	usecase    usecase.CategoryUsecase
	baseLogger *zap.Logger
}

func NewCategoriesHandler(categoryRepo domain.CategoryRepo, baseLogger *zap.Logger) *CategoriesHandler {
	return &CategoriesHandler{
		usecase:    usecase.NewCategoryUsecase(categoryRepo),
		baseLogger: baseLogger,
	}
}

func (h *CategoriesHandler) logger(ctx context.Context) *zap.Logger {
	logger := ctx.Value(common.LoggerContextKey).(*zap.Logger)
	if logger != nil {
		return logger
	}
	return h.baseLogger
}
