package usecase

import (
	orderDomain "vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/products/domain"
)

type PostProductReviewInDTO struct {
	ProductID int64
	BuyerID   int64
	Review    *string
	Rating    int
}

type PostProductReviewOutDTO struct {
}

func (u *ProductReviewUsecaseImpl) PostProductReview(dto PostProductReviewInDTO) (PostProductReviewOutDTO, error) {
	orders, err := u.orderRepo.SelectByProductIDAndBuyerID(dto.ProductID, dto.BuyerID)
	if err != nil {
		return PostProductReviewOutDTO{}, err
	}

	for _, order := range orders {
		if order.OrderStatusID == domain.OrderStatusIDCanPostReview {
			err := u.productReviewRepo.Insert(domain.ProductReview{
				BuyerID: dto.BuyerID, ProductID: dto.ProductID, Review: dto.Review, Rating: dto.Rating,
			})
			return PostProductReviewOutDTO{}, err
		}
	}

	return PostProductReviewOutDTO{}, orderDomain.ErrOrdersNotFound
}
