package usecase

import (
	"context"
	"errors"
	"testing"
	"time"
	ordersDomain "vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/products/domain"
)

type reviewReadRepoStub struct {
	gotParams domain.ProductReviewsParams
	page      domain.ProductReviewsPage
	review    domain.ProductReview
	updated   domain.ProductReview
	err       error
}

func (r *reviewReadRepoStub) Insert(_ domain.ProductReview) error {
	return nil
}

func (r *reviewReadRepoStub) SelectByProductID(_ context.Context, params domain.ProductReviewsParams) (domain.ProductReviewsPage, error) {
	r.gotParams = params
	return r.page, r.err
}

func (r *reviewReadRepoStub) SelectByProductIDAndBuyerID(_ context.Context, _, _ int64) (domain.ProductReview, error) {
	return r.review, r.err
}

func (r *reviewReadRepoStub) Update(_ context.Context, pr domain.ProductReview) (domain.ProductReview, error) {
	r.updated = pr
	return pr, r.err
}

func TestGetProductReviews(t *testing.T) {
	now := time.Date(2026, 11, 24, 15, 0, 0, 0, time.UTC)

	t.Run("defaults sortBy and limit", func(t *testing.T) {
		readRepo := &reviewReadRepoStub{
			page: domain.ProductReviewsPage{
				Reviews: []domain.ProductReviewListItem{{BuyerID: 1, BuyerName: "Иван", Rating: 5, CreatedAt: now}},
			},
		}
		u := NewProductReviewUsecase(readRepo, nil)

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
		u := NewProductReviewUsecase(readRepo, nil)

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
		u := NewProductReviewUsecase(&reviewReadRepoStub{}, nil)

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
		u := NewProductReviewUsecase(&reviewReadRepoStub{err: wantErr}, nil)

		_, err := u.GetProductReviews(GetProductReviewsInDTO{ProductID: 42})
		if !errors.Is(err, wantErr) {
			t.Fatalf("got error %v, want %v", err, wantErr)
		}
	})
}

func TestGetProductReviewUsecase(t *testing.T) {
	text := "Норм"
	t.Run("success", func(t *testing.T) {
		readRepo := &reviewReadRepoStub{
			review: domain.ProductReview{Rating: 4, Review: &text},
		}
		u := NewProductReviewUsecase(readRepo, nil)
		out, err := u.GetProductReview(GetProductReviewInDTO{ProductID: 12, BuyerID: 34})
		if err != nil || out.Rating != 4 || out.Review == nil || *out.Review != text {
			t.Fatalf("unexpected result: out=%+v err=%v", out, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := NewProductReviewUsecase(&reviewReadRepoStub{err: domain.ErrProductReviewNotFound}, nil)
		_, err := u.GetProductReview(GetProductReviewInDTO{ProductID: 12, BuyerID: 34})
		if !errors.Is(err, domain.ErrProductReviewNotFound) {
			t.Fatalf("got error %v, want %v", err, domain.ErrProductReviewNotFound)
		}
	})
}

func TestPatchProductReviewUsecase(t *testing.T) {
	t.Run("invalid rating", func(t *testing.T) {
		u := NewProductReviewUsecase(&reviewReadRepoStub{}, nil)
		_, err := u.PatchProductReview(PatchProductReviewInDTO{ProductID: 12, BuyerID: 34, Rating: 0})
		if !errors.Is(err, domain.ErrProductReviewRatingInvalid) {
			t.Fatalf("got error %v, want %v", err, domain.ErrProductReviewRatingInvalid)
		}
	})

	t.Run("order not received", func(t *testing.T) {
		u := NewProductReviewUsecase(&reviewReadRepoStub{}, reviewOrderRepo{orders: []ordersDomain.Order{{OrderStatusID: 1}}})
		_, err := u.PatchProductReview(PatchProductReviewInDTO{ProductID: 12, BuyerID: 34, Rating: 5})
		if !errors.Is(err, ordersDomain.ErrOrdersNotFound) {
			t.Fatalf("got error %v, want %v", err, ordersDomain.ErrOrdersNotFound)
		}
	})

	t.Run("success trims empty review to nil", func(t *testing.T) {
		repo := &reviewReadRepoStub{}
		u := NewProductReviewUsecase(repo, reviewOrderRepo{orders: []ordersDomain.Order{{OrderStatusID: 4}}})
		empty := "   "
		out, err := u.PatchProductReview(PatchProductReviewInDTO{ProductID: 12, BuyerID: 34, Rating: 5, Review: &empty})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.updated.Review != nil || out.Rating != 5 {
			t.Fatalf("expected nil review and rating 5, got updated=%+v out=%+v", repo.updated, out)
		}
	})
}
