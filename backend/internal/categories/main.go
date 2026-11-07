package categories

import (
	"net/http"
	"vibe_market/backend/internal/categories/delivery"
	"vibe_market/backend/internal/categories/domain"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

func NewServeMux(categoryRepo domain.CategoryRepo, baseLogger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	categoriesHandler := delivery.NewCategoriesHandler(categoryRepo, baseLogger)

	mux.Handle("/", common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(categoriesHandler.Categories)))

	return mux
}
