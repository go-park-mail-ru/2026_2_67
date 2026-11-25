package usecase

import (
	"context"
	"strings"
	orderDomain "vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/products/domain"
)

const (
	defaultReviewsLimit = 10
	maxReviewsLimit     = 50
)

type PostProductReviewInDTO struct {
	ProductID int64
	BuyerID   int64
	Review    *string
	Rating    int
}

type PostProductReviewOutDTO struct {
}

type GetProductReviewsInDTO struct {
	Context   context.Context
	ProductID int64
	SortBy    domain.ProductReviewSortBy
	Limit     int
	Cursor    *domain.ProductReviewCursor
}

type GetProductReviewsOutDTO struct {
	Reviews    []domain.ProductReviewListItem
	NextCursor *domain.ProductReviewCursor
}

func (u *ProductReviewUsecaseImpl) PostProductReview(dto PostProductReviewInDTO) (PostProductReviewOutDTO, error) {
	orders, err := u.orderRepo.SelectByProductIDAndBuyerID(dto.ProductID, dto.BuyerID)
	if err != nil {
		return PostProductReviewOutDTO{}, err
	}

	*dto.Review = strings.TrimSpace(*dto.Review)
	if *dto.Review == "" {
		dto.Review = nil
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

func isValidReviewSortBy(sortBy domain.ProductReviewSortBy) bool {
	switch sortBy {
	case domain.ReviewSortByDescDate, domain.ReviewSortByAscDate, domain.ReviewSortByDescRating, domain.ReviewSortByAscRating:
		return true
	default:
		return false
	}
}

func (u *ProductReviewUsecaseImpl) GetProductReviews(dto GetProductReviewsInDTO) (GetProductReviewsOutDTO, error) {
	sortBy := dto.SortBy
	if sortBy == "" {
		sortBy = domain.ReviewSortByDescDate
	}
	if !isValidReviewSortBy(sortBy) {
		return GetProductReviewsOutDTO{}, domain.ErrProductReviewSortByInvalid
	}

	limit := dto.Limit
	if limit <= 0 {
		limit = defaultReviewsLimit
	} else if limit > maxReviewsLimit {
		limit = maxReviewsLimit
	}

	ctx := dto.Context
	if ctx == nil {
		ctx = context.Background()
	}

	page, err := u.productReviewReadRepo.SelectByProductID(ctx, domain.ProductReviewsParams{
		ProductID: dto.ProductID,
		SortBy:    sortBy,
		Limit:     limit,
		Cursor:    dto.Cursor,
	})
	if err != nil {
		return GetProductReviewsOutDTO{}, err
	}

	return GetProductReviewsOutDTO{
		Reviews:    page.Reviews,
		NextCursor: page.NextCursor,
	}, nil
}
