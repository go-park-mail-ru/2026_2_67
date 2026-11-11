package usecase

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"vibe_market/backend/internal/orders/domain"
)

type listOrdersRepoStub struct {
	t       *testing.T
	orders  []domain.OrderDetails
	err     error
	calls   int
	ctx     context.Context
	buyerID int64
}

var _ domain.OrderRepo = (*listOrdersRepoStub)(nil)

func (s *listOrdersRepoStub) ListByBuyer(ctx context.Context, buyerID int64) ([]domain.OrderDetails, error) {
	s.calls++
	s.ctx, s.buyerID = ctx, buyerID
	return s.orders, s.err
}

func (s *listOrdersRepoStub) CreateOrderTx(context.Context, int64, int64, string) (int64, []domain.ProductStockError, error) {
	s.t.Fatal("unexpected CreateOrderTx call")
	return 0, nil, nil
}

func assertListOrdersCall(t *testing.T, repo *listOrdersRepoStub, ctx context.Context, buyerID int64) {
	t.Helper()
	if repo.calls != 1 || repo.ctx != ctx || repo.buyerID != buyerID {
		t.Fatalf("ListByBuyer calls = %d, context = %v, buyerID = %d; want one call with original context and buyerID %d", repo.calls, repo.ctx, repo.buyerID, buyerID)
	}
}

func TestListOrdersUsecaseMappingAndTotals(t *testing.T) {
	created := time.Date(2026, 10, 27, 10, 15, 0, 123, time.FixedZone("UTC+3", 3*60*60))
	updated := created.Add(time.Hour)
	description := "product description"
	categoryID := int64(8)
	repo := &listOrdersRepoStub{t: t, orders: []domain.OrderDetails{
		{
			ID: 90, Status: domain.OrderStatus{ID: 2, Name: "paid"}, CreatedAt: created, UpdatedAt: updated,
			PickupPoint: domain.OrderPickupPoint{ID: 12, Longitude: 37.6, Latitude: 55.7, StartTime: "09:00", EndTime: "21:00"},
			Promocode:   &domain.OrderPromocode{ID: 4, Code: "SALE", Discount: 25, AvailableCategoryID: &categoryID, StartDatetime: created.Add(-time.Hour), EndDatetime: updated.Add(time.Hour)},
			Products: []domain.OrderProductDetails{
				{ID: 15, ProductID: 33, Name: "Book", Description: &description, Category: domain.OrderCategory{ID: 8, Name: "Books"}, Seller: domain.OrderSeller{ID: 7, Name: "Shop"}, PictureURLs: []string{"first.jpg", "second.jpg"}, Count: 3, Price: 4_000_000_000},
				{ID: 11, ProductID: 22, Name: "Pen", Category: domain.OrderCategory{ID: 9, Name: "Stationery"}, Seller: domain.OrderSeller{ID: 6, Name: "Pens"}, PictureURLs: []string{}, Count: 2, Price: 125},
			},
			Notifications: []domain.OrderNotification{
				{ID: 20, Status: domain.OrderStatus{ID: 1, Name: "created"}, IsRead: true, CreatedAt: created},
				{ID: 21, Status: domain.OrderStatus{ID: 2, Name: "paid"}, CreatedAt: updated},
			},
		},
		{
			ID: 80, Status: domain.OrderStatus{ID: 1, Name: "created"}, CreatedAt: updated, UpdatedAt: updated,
			Products: []domain.OrderProductDetails{{ID: 10, ProductID: 14, Name: "Other", Count: 1, Price: 50}},
		},
	}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	got, err := ListOrdersUsecase(ctx, 42, repo)
	if err != nil {
		t.Fatal(err)
	}
	assertListOrdersCall(t, repo, ctx, 42)
	want := ListOrdersOutDTO{Orders: []OrderDTO{
		{
			ID: 90, Status: OrderNamedEntityDTO{ID: 2, Name: "paid"}, CreatedAt: created.UTC(), UpdatedAt: updated.UTC(),
			PickupPoint: OrderPickupPointDTO{ID: 12, Longitude: 37.6, Latitude: 55.7, StartTime: "09:00", EndTime: "21:00"},
			Promocode:   &OrderPromocodeDTO{ID: 4, Code: "SALE", Discount: 25, AvailableCategoryID: &categoryID, StartDatetime: created.Add(-time.Hour).UTC(), EndDatetime: updated.Add(time.Hour).UTC()},
			Products: []OrderProductDTO{
				{ID: 15, ProductID: 33, Name: "Book", Description: &description, Category: OrderNamedEntityDTO{ID: 8, Name: "Books"}, Seller: OrderNamedEntityDTO{ID: 7, Name: "Shop"}, PictureURLs: []string{"first.jpg", "second.jpg"}, Count: 3, UnitPrice: 4_000_000_000, TotalPrice: 12_000_000_000},
				{ID: 11, ProductID: 22, Name: "Pen", Category: OrderNamedEntityDTO{ID: 9, Name: "Stationery"}, Seller: OrderNamedEntityDTO{ID: 6, Name: "Pens"}, PictureURLs: []string{}, Count: 2, UnitPrice: 125, TotalPrice: 250},
			},
			Notifications: []OrderNotificationDTO{
				{ID: 20, Status: OrderNamedEntityDTO{ID: 1, Name: "created"}, IsRead: true, CreatedAt: created.UTC()},
				{ID: 21, Status: OrderNamedEntityDTO{ID: 2, Name: "paid"}, CreatedAt: updated.UTC()},
			},
			TotalCount: 5, TotalPrice: 12_000_000_250,
		},
		{
			ID: 80, Status: OrderNamedEntityDTO{ID: 1, Name: "created"}, CreatedAt: updated.UTC(), UpdatedAt: updated.UTC(),
			Products:      []OrderProductDTO{{ID: 10, ProductID: 14, Name: "Other", PictureURLs: []string{}, Count: 1, UnitPrice: 50, TotalPrice: 50}},
			Notifications: []OrderNotificationDTO{}, TotalCount: 1, TotalPrice: 50,
		},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ListOrdersUsecase() = %#v, want %#v", got, want)
	}
	got.Orders[0].Products[0].PictureURLs[0] = "changed.jpg"
	if repo.orders[0].Products[0].PictureURLs[0] != "first.jpg" {
		t.Error("output pictures share backing storage with repository data")
	}
	if repo.orders[0].CreatedAt.Location() == time.UTC {
		t.Error("repository timestamp was mutated")
	}
}

func TestListOrdersUsecaseEmptyResults(t *testing.T) {
	for _, orders := range [][]domain.OrderDetails{nil, {}} {
		name := "nil"
		if orders != nil {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			repo := &listOrdersRepoStub{t: t, orders: orders}
			ctx := context.Background()
			got, err := ListOrdersUsecase(ctx, 7, repo)
			if err != nil {
				t.Fatal(err)
			}
			assertListOrdersCall(t, repo, ctx, 7)
			if !reflect.DeepEqual(got, ListOrdersOutDTO{Orders: []OrderDTO{}}) {
				t.Fatalf("result = %#v, want non-nil empty orders", got)
			}
		})
	}
}

func TestListOrdersUsecaseEmptyNestedCollections(t *testing.T) {
	for _, empty := range []bool{false, true} {
		t.Run(fmt.Sprintf("empty=%t", empty), func(t *testing.T) {
			order := domain.OrderDetails{ID: 1, Promocode: &domain.OrderPromocode{ID: 2}}
			if empty {
				order.Products = []domain.OrderProductDetails{}
				order.Notifications = []domain.OrderNotification{}
			}
			repo := &listOrdersRepoStub{t: t, orders: []domain.OrderDetails{order}}
			got, err := ListOrdersUsecase(context.Background(), 7, repo)
			if err != nil {
				t.Fatal(err)
			}
			want := ListOrdersOutDTO{Orders: []OrderDTO{{ID: 1, Promocode: &OrderPromocodeDTO{ID: 2}, Products: []OrderProductDTO{}, Notifications: []OrderNotificationDTO{}}}}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("result = %#v, want %#v", got, want)
			}
		})
	}
}

func TestListOrdersUsecaseRepositoryErrors(t *testing.T) {
	repoErr := errors.New("repository unavailable")
	for _, err := range []error{repoErr, fmt.Errorf("list: %w", repoErr), context.Canceled, context.DeadlineExceeded} {
		t.Run(err.Error(), func(t *testing.T) {
			repo := &listOrdersRepoStub{t: t, orders: []domain.OrderDetails{{ID: 1}}, err: err}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			got, gotErr := ListOrdersUsecase(ctx, 42, repo)
			assertListOrdersCall(t, repo, ctx, 42)
			if gotErr != err {
				t.Fatalf("error = %v, want original error %v", gotErr, err)
			}
			if !reflect.DeepEqual(got, ListOrdersOutDTO{}) {
				t.Fatalf("result = %#v, want zero DTO without partial data", got)
			}
		})
	}
}
