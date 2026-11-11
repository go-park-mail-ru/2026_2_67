package usecase

import (
	"context"
	"errors"
	"testing"

	"vibe_market/backend/internal/orders/domain"
)

type createOrderRepoStub struct {
	createOrderTx func(ctx context.Context, buyerID, pickupPointID int64, promoCode string) (int64, []domain.ProductStockError, error)
	listByBuyer   func(ctx context.Context, buyerID int64) ([]domain.OrderDetails, error)
}

var _ domain.OrderRepo = (*createOrderRepoStub)(nil)

func (s *createOrderRepoStub) CreateOrderTx(ctx context.Context, buyerID, pickupPointID int64, promoCode string) (int64, []domain.ProductStockError, error) {
	if s.createOrderTx != nil {
		return s.createOrderTx(ctx, buyerID, pickupPointID, promoCode)
	}
	panic("unexpected CreateOrderTx call")
}

func (s *createOrderRepoStub) ListByBuyer(ctx context.Context, buyerID int64) ([]domain.OrderDetails, error) {
	if s.listByBuyer != nil {
		return s.listByBuyer(ctx, buyerID)
	}
	panic("unexpected ListByBuyer call")
}

func TestCreateOrderUsecase(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid pickup point ID", func(t *testing.T) {
		repo := &createOrderRepoStub{}
		for _, id := range []int64{0, -1, -100} {
			out, valErrs, stockErrs, err := CreateOrderUsecase(ctx, CreateOrderInDTO{
				BuyerID:       10,
				PickupPointID: id,
			}, repo)

			if !errors.Is(err, domain.ErrPickupPointInvalid) {
				t.Fatalf("id %d: got err %v, want %v", id, err, domain.ErrPickupPointInvalid)
			}
			if valErrs.PickupPointIDErrMessage == "" {
				t.Errorf("expected validation message for pickup point ID %d", id)
			}
			if out.OrderID != 0 || len(stockErrs) != 0 {
				t.Errorf("unexpected out or stockErrs: %+v, %+v", out, stockErrs)
			}
		}
	})

	t.Run("pickup point not found from repo", func(t *testing.T) {
		repo := &createOrderRepoStub{
			createOrderTx: func(ctx context.Context, buyerID, pickupPointID int64, promoCode string) (int64, []domain.ProductStockError, error) {
				return 0, nil, domain.ErrPickupPointInvalid
			},
		}

		_, valErrs, _, err := CreateOrderUsecase(ctx, CreateOrderInDTO{
			BuyerID:       10,
			PickupPointID: 99,
		}, repo)

		if !errors.Is(err, domain.ErrPickupPointInvalid) {
			t.Fatalf("got err %v, want %v", err, domain.ErrPickupPointInvalid)
		}
		if valErrs.PickupPointIDErrMessage == "" {
			t.Error("expected PickupPointIDErrMessage")
		}
	})

	t.Run("invalid promocode from repo", func(t *testing.T) {
		repo := &createOrderRepoStub{
			createOrderTx: func(ctx context.Context, buyerID, pickupPointID int64, promoCode string) (int64, []domain.ProductStockError, error) {
				return 0, nil, domain.ErrPromocodeInvalid
			},
		}

		_, valErrs, _, err := CreateOrderUsecase(ctx, CreateOrderInDTO{
			BuyerID:       10,
			PickupPointID: 5,
			Promocode:     "INVALID",
		}, repo)

		if !errors.Is(err, domain.ErrPromocodeInvalid) {
			t.Fatalf("got err %v, want %v", err, domain.ErrPromocodeInvalid)
		}
		if valErrs.PromocodeErrMessage == "" {
			t.Error("expected PromocodeErrMessage")
		}
	})

	t.Run("insufficient stock errors", func(t *testing.T) {
		repo := &createOrderRepoStub{
			createOrderTx: func(ctx context.Context, buyerID, pickupPointID int64, promoCode string) (int64, []domain.ProductStockError, error) {
				return 0, []domain.ProductStockError{
					{ProductID: 1, Name: "Item", Seller: "Seller", ErrMessage: "not enough stock"},
				}, domain.ErrOrderStockInsufficient
			},
		}

		_, _, stockErrs, err := CreateOrderUsecase(ctx, CreateOrderInDTO{
			BuyerID:       10,
			PickupPointID: 5,
		}, repo)

		if !errors.Is(err, domain.ErrOrderStockInsufficient) {
			t.Fatalf("got err %v, want %v", err, domain.ErrOrderStockInsufficient)
		}
		if len(stockErrs) != 1 || stockErrs[0].ProductID != 1 {
			t.Fatalf("unexpected stock errors: %+v", stockErrs)
		}
	})

	t.Run("database internal error", func(t *testing.T) {
		expectedErr := errors.New("db error")
		repo := &createOrderRepoStub{
			createOrderTx: func(ctx context.Context, buyerID, pickupPointID int64, promoCode string) (int64, []domain.ProductStockError, error) {
				return 0, nil, expectedErr
			},
		}

		_, _, _, err := CreateOrderUsecase(ctx, CreateOrderInDTO{
			BuyerID:       10,
			PickupPointID: 5,
		}, repo)

		if !errors.Is(err, expectedErr) {
			t.Fatalf("got err %v, want %v", err, expectedErr)
		}
	})

	t.Run("success", func(t *testing.T) {
		repo := &createOrderRepoStub{
			createOrderTx: func(ctx context.Context, buyerID, pickupPointID int64, promoCode string) (int64, []domain.ProductStockError, error) {
				if buyerID != 10 || pickupPointID != 5 || promoCode != "SALE" {
					t.Errorf("unexpected args: buyer=%d, pickup=%d, promo=%s", buyerID, pickupPointID, promoCode)
				}
				return 777, nil, nil
			},
		}

		out, valErrs, stockErrs, err := CreateOrderUsecase(ctx, CreateOrderInDTO{
			BuyerID:       10,
			PickupPointID: 5,
			Promocode:     "  SALE  ", // should be trimmed
		}, repo)

		if err != nil || valErrs.HasErrors() || len(stockErrs) != 0 {
			t.Fatalf("unexpected result: err=%v, valErrs=%+v, stockErrs=%+v", err, valErrs, stockErrs)
		}
		if out.OrderID != 777 {
			t.Errorf("got OrderID %d, want 777", out.OrderID)
		}
	})
}
