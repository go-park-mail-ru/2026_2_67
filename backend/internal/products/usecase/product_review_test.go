package usecase

import (
	"context"
	"errors"
	"testing"

	ordersDomain "vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/products/domain"
)

type reviewOrderRepo struct {
	ordersDomain.OrderRepo
	err    error
	orders []ordersDomain.Order
}

func (r reviewOrderRepo) SelectByProductIDAndBuyerID(productID, buyerID int64) ([]ordersDomain.Order, error) {
	if productID != 12 || buyerID != 34 {
		panic("unexpected product or buyer ID")
	}
	return r.orders, r.err
}

type reviewRepoStub struct {
	reviews []domain.ProductReview
	err     error
}

func (r *reviewRepoStub) Insert(review domain.ProductReview) error {
	r.reviews = append(r.reviews, review)
	return r.err
}

func (r *reviewRepoStub) SelectByProductID(_ context.Context, _ domain.ProductReviewsParams) (domain.ProductReviewsPage, error) {
	return domain.ProductReviewsPage{}, nil
}

func (r *reviewRepoStub) SelectByProductIDAndBuyerID(_ context.Context, _, _ int64) (domain.ProductReview, error) {
	return domain.ProductReview{}, nil
}

func (r *reviewRepoStub) Update(_ context.Context, pr domain.ProductReview) (domain.ProductReview, error) {
	return pr, nil
}

func TestProductReviewPostUsesOrderRepository(t *testing.T) {
	wantErr := errors.New("order lookup failed")
	u := NewProductReviewUsecase(nil, reviewOrderRepo{err: wantErr})
	_, err := u.Post(domain.PostProductReviewInDTO{ProductID: 12, BuyerID: 34})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got error %v, want %v", err, wantErr)
	}
}

func TestProductReviewPost(t *testing.T) {
	reviewText := "Good"
	tests := []struct {
		name   string
		orders []ordersDomain.Order
		insert bool
	}{
		{name: "no orders"},
		{name: "order not received", orders: []ordersDomain.Order{{OrderStatusID: 1}}},
		{name: "received order", orders: []ordersDomain.Order{{OrderStatusID: 1}, {OrderStatusID: 4}}, insert: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &reviewRepoStub{}
			u := NewProductReviewUsecase(repo, reviewOrderRepo{orders: tt.orders})
			_, err := u.Post(domain.PostProductReviewInDTO{ProductID: 12, BuyerID: 34, Review: &reviewText, Rating: 5})
			if !tt.insert {
				if !errors.Is(err, ordersDomain.ErrOrdersNotFound) || len(repo.reviews) != 0 {
					t.Fatalf("err=%v, reviews=%+v", err, repo.reviews)
				}
				return
			}
			if err != nil || len(repo.reviews) != 1 || repo.reviews[0].Rating != 5 || repo.reviews[0].BuyerID != 34 || repo.reviews[0].ProductID != 12 || repo.reviews[0].Review != &reviewText {
				t.Fatalf("err=%v, reviews=%+v", err, repo.reviews)
			}
		})
	}
}
