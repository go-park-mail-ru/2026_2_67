package usecase

import (
	orderDomain "vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/products/domain"
)

type ProductUsecase interface {
	Product(id int64) (ProductOutDTO, error)
	Products(dto domain.ProductsSearchParams) (ProductsOutDTO, error)
}

type ProductReviewUsecase interface {
	PostProductReview(dto PostProductReviewInDTO) (PostProductReviewOutDTO, error)
}

type ProductUsecaseImpl struct {
	productRepo domain.ProductRepo
}

type ProductReviewUsecaseImpl struct {
	productReviewRepo domain.ProductReviewRepo
	orderRepo         orderDomain.OrderRepo
}

func NewProductUsecase(productRepo domain.ProductRepo) *ProductUsecaseImpl {
	return &ProductUsecaseImpl{productRepo: productRepo}
}

func NewProductReviewUsecase(productReviewRepo domain.ProductReviewRepo, orderRepo orderDomain.OrderRepo) *ProductReviewUsecaseImpl {
	return &ProductReviewUsecaseImpl{productReviewRepo: productReviewRepo, orderRepo: orderRepo}
}
