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

type GetProductReviewInDTO struct {
	Context   context.Context
	ProductID int64
	BuyerID   int64
}

type GetProductReviewOutDTO struct {
	Rating int
	Review *string
}

type PatchProductReviewInDTO struct {
	Context   context.Context
	ProductID int64
	BuyerID   int64
	Review    *string
	Rating    int
}

type PatchProductReviewOutDTO struct {
	Rating int
	Review *string
}

func normalizeReviewText(review *string) *string {
	if review == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*review)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (u *ProductReviewUsecaseImpl) ensureBuyerCanReview(productID, buyerID int64) error {
	orders, err := u.orderRepo.SelectByProductIDAndBuyerID(productID, buyerID)
	if err != nil {
		return err
	}
	for _, order := range orders {
		if order.OrderStatusID == domain.OrderStatusIDCanPostReview {
			return nil
		}
	}
	return orderDomain.ErrOrdersNotFound
}

func (u *ProductReviewUsecaseImpl) PostProductReview(dto PostProductReviewInDTO) (PostProductReviewOutDTO, error) {
	if err := u.ensureBuyerCanReview(dto.ProductID, dto.BuyerID); err != nil {
		return PostProductReviewOutDTO{}, err
	}

	err := u.productReviewRepo.Insert(domain.ProductReview{
		BuyerID:   dto.BuyerID,
		ProductID: dto.ProductID,
		Review:    normalizeReviewText(dto.Review),
		Rating:    dto.Rating,
	})
	return PostProductReviewOutDTO{}, err
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

	page, err := u.productReviewRepo.SelectByProductID(ctx, domain.ProductReviewsParams{
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

func (u *ProductReviewUsecaseImpl) GetProductReview(dto GetProductReviewInDTO) (GetProductReviewOutDTO, error) {
	ctx := dto.Context
	if ctx == nil {
		ctx = context.Background()
	}

	pr, err := u.productReviewRepo.SelectByProductIDAndBuyerID(ctx, dto.ProductID, dto.BuyerID)
	if err != nil {
		return GetProductReviewOutDTO{}, err
	}

	return GetProductReviewOutDTO{
		Rating: pr.Rating,
		Review: pr.Review,
	}, nil
}

func (u *ProductReviewUsecaseImpl) PatchProductReview(dto PatchProductReviewInDTO) (PatchProductReviewOutDTO, error) {
	if dto.Rating < 1 || dto.Rating > 5 {
		return PatchProductReviewOutDTO{}, domain.ErrProductReviewRatingInvalid
	}

	if err := u.ensureBuyerCanReview(dto.ProductID, dto.BuyerID); err != nil {
		return PatchProductReviewOutDTO{}, err
	}

	ctx := dto.Context
	if ctx == nil {
		ctx = context.Background()
	}

	updated, err := u.productReviewRepo.Update(ctx, domain.ProductReview{
		BuyerID:   dto.BuyerID,
		ProductID: dto.ProductID,
		Review:    normalizeReviewText(dto.Review),
		Rating:    dto.Rating,
	})
	if err != nil {
		return PatchProductReviewOutDTO{}, err
	}

	return PatchProductReviewOutDTO{
		Rating: updated.Rating,
		Review: updated.Review,
	}, nil
}
