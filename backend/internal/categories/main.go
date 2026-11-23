package categories

import (
	"net/http"
	"vibe_market/backend/internal/categories/delivery"
	"vibe_market/backend/internal/categories/domain"
	"vibe_market/backend/internal/categories/usecase"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

func NewServeMux(categoryRepo domain.CategoryRepo, baseLogger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo)
	categoriesHandler := delivery.NewCategoriesHandler(categoryRepo, categoryUsecase, baseLogger)

	handler := common.RequireHTTPMethodMiddleware(http.MethodGet)(http.HandlerFunc(categoriesHandler.Categories))
	mux.Handle("/", handler)

	return mux
}
