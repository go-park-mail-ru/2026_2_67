package usecase

import (
	orderDomain "vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/products/domain"
)

type ProductUsecase struct {
	productRepo domain.ProductRepo
}

type ProductReviewUsecase struct {
	productReviewRepo domain.ProductReviewRepo
	orderRepo         orderDomain.OrderRepo
}

func NewProductUsecase(productRepo domain.ProductRepo) *ProductUsecase {
	return &ProductUsecase{productRepo: productRepo}
}

func NewProductReviewUsecase(productReviewRepo domain.ProductReviewRepo, orderRepo orderDomain.OrderRepo) *ProductReviewUsecase {
	return &ProductReviewUsecase{productReviewRepo: productReviewRepo, orderRepo: orderRepo}
}
