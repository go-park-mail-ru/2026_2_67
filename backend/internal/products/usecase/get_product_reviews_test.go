package usecase

import (
	"context"
	"errors"
	"testing"
	"time"
	"vibe_market/backend/internal/products/domain"
)

type reviewReadRepoStub struct {
	gotParams domain.ProductReviewsParams
	page      domain.ProductReviewsPage
	err       error
}

func (r *reviewReadRepoStub) SelectByProductID(_ context.Context, params domain.ProductReviewsParams) (domain.ProductReviewsPage, error) {
	r.gotParams = params
	return r.page, r.err
}

func TestGetProductReviews(t *testing.T) {
	now := time.Date(2026, 11, 24, 15, 0, 0, 0, time.UTC)

	t.Run("defaults sortBy and limit", func(t *testing.T) {
		readRepo := &reviewReadRepoStub{
			page: domain.ProductReviewsPage{
				Reviews: []domain.ProductReviewListItem{{BuyerID: 1, BuyerName: "Иван", Rating: 5, CreatedAt: now}},
			},
		}
		u := NewProductReviewUsecaseWithReader(nil, readRepo, nil)

		out, err := u.GetProductReviews(GetProductReviewsInDTO{ProductID: 42})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if readRepo.gotParams.SortBy != domain.ReviewSortByDescDate {
			t.Fatalf("got sortBy %q, want %q", readRepo.gotParams.SortBy, domain.ReviewSortByDescDate)
		}
		if readRepo.gotParams.Limit != defaultReviewsLimit {
			t.Fatalf("got limit %d, want %d", readRepo.gotParams.Limit, defaultReviewsLimit)
		}
		if len(out.Reviews) != 1 || out.Reviews[0].BuyerName != "Иван" {
			t.Fatalf("unexpected output: %+v", out)
		}
	})

	t.Run("clamps max limit", func(t *testing.T) {
		readRepo := &reviewReadRepoStub{}
		u := NewProductReviewUsecaseWithReader(nil, readRepo, nil)

		_, err := u.GetProductReviews(GetProductReviewsInDTO{
			ProductID: 42,
			SortBy:    domain.ReviewSortByAscRating,
			Limit:     500,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if readRepo.gotParams.Limit != maxReviewsLimit {
			t.Fatalf("got limit %d, want %d", readRepo.gotParams.Limit, maxReviewsLimit)
		}
	})

	t.Run("invalid sortBy", func(t *testing.T) {
		u := NewProductReviewUsecaseWithReader(nil, &reviewReadRepoStub{}, nil)

		_, err := u.GetProductReviews(GetProductReviewsInDTO{
			ProductID: 42,
			SortBy:    "unknown-sort",
		})
		if !errors.Is(err, domain.ErrProductReviewSortByInvalid) {
			t.Fatalf("got error %v, want %v", err, domain.ErrProductReviewSortByInvalid)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		wantErr := errors.New("db error")
		u := NewProductReviewUsecaseWithReader(nil, &reviewReadRepoStub{err: wantErr}, nil)

		_, err := u.GetProductReviews(GetProductReviewsInDTO{ProductID: 42})
		if !errors.Is(err, wantErr) {
			t.Fatalf("got error %v, want %v", err, wantErr)
		}
	})
}
