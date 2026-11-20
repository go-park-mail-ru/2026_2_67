package usecase

import (
	ordersDomain "vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/products/domain"
)

func (u *ProductReviewUsecase) Post(dto domain.PostProductReviewInDTO) (domain.PostProductReviewOutDTO, error) {
	orders, err := u.orderRepo.SelectByProductIDAndBuyerID(dto.ProductID, dto.BuyerID)
	if err != nil {
		return domain.PostProductReviewOutDTO{}, err
	}

	for _, order := range orders {
		if order.OrderStatusID == domain.OrderStatusIDCanPostReview {
			err := u.productReviewRepo.Insert(domain.ProductReview{
				BuyerID: dto.BuyerID, ProductID: dto.ProductID, Review: dto.Review, Rating: dto.Rating,
			})
			return domain.PostProductReviewOutDTO{}, err
		}
	}

	return domain.PostProductReviewOutDTO{}, ordersDomain.ErrOrdersNotFound
}
