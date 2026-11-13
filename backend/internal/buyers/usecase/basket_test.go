package usecase

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"vibe_market/backend/internal/buyers/domain"
)

type basketRepoStub struct {
	getBasket             func(buyerID int64) (domain.BasketSummary, error)
	addProduct            func(buyerID int64, productID int64, count int) error
	updateProductCount    func(buyerID int64, productID int64, count int) error
	getNearestPickupPoint func(lat, lng float64) (domain.PickupPoint, error)
	getPromocode          func(code string) (*domain.Promocode, error)
}

var _ domain.BasketRepo = (*basketRepoStub)(nil)

func (s *basketRepoStub) GetBasket(buyerID int64) (domain.BasketSummary, error) {
	if s.getBasket != nil {
		return s.getBasket(buyerID)
	}
	panic("unexpected GetBasket call")
}

func (s *basketRepoStub) AddProduct(buyerID int64, productID int64, count int) error {
	if s.addProduct != nil {
		return s.addProduct(buyerID, productID, count)
	}
	panic("unexpected AddProduct call")
}

func (s *basketRepoStub) UpdateProductCount(buyerID int64, productID int64, count int) error {
	if s.updateProductCount != nil {
		return s.updateProductCount(buyerID, productID, count)
	}
	panic("unexpected UpdateProductCount call")
}

func (s *basketRepoStub) GetNearestPickupPoint(lat, lng float64) (domain.PickupPoint, error) {
	if s.getNearestPickupPoint != nil {
		return s.getNearestPickupPoint(lat, lng)
	}
	panic("unexpected GetNearestPickupPoint call")
}

func (s *basketRepoStub) GetPromocode(code string) (*domain.Promocode, error) {
	if s.getPromocode != nil {
		return s.getPromocode(code)
	}
	panic("unexpected GetPromocode call")
}

func TestGetBasketUsecase(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &basketRepoStub{
			getBasket: func(buyerID int64) (domain.BasketSummary, error) {
				if buyerID != 10 {
					t.Fatalf("expected buyerID 10, got %d", buyerID)
				}
				return domain.BasketSummary{
					Items: []domain.BasketItem{
						{
							ProductID:    1,
							CategoryName: "Electronics",
							Name:         "Phone",
							PictureURL:   "/phone.jpg",
							Price:        1000,
							Discount:     10,
							PriceResult:  900,
							Count:        2,
						},
					},
					BasketSum:       2000,
					BasketSumResult: 1800,
				}, nil
			},
		}

		out, err := GetBasketUsecase(10, repo)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.BasketSum != 2000 || out.BasketSumResult != 1800 {
			t.Errorf("unexpected sums: %+v", out)
		}
		if len(out.BasketProducts) != 1 {
			t.Fatalf("expected 1 product, got %d", len(out.BasketProducts))
		}
		expectedItem := BasketProductItemDTO{
			ID:          1,
			Category:    "Electronics",
			Name:        "Phone",
			PictureURL:  "/phone.jpg",
			Price:       1000,
			Discount:    10,
			PriceResult: 900,
			Count:       2,
		}
		if !reflect.DeepEqual(out.BasketProducts[0], expectedItem) {
			t.Errorf("item mismatch: got %+v, want %+v", out.BasketProducts[0], expectedItem)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("db error")
		repo := &basketRepoStub{
			getBasket: func(buyerID int64) (domain.BasketSummary, error) {
				return domain.BasketSummary{}, expectedErr
			},
		}

		_, err := GetBasketUsecase(10, repo)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("got err %v, want %v", err, expectedErr)
		}
	})
}

func TestAddProductUsecase(t *testing.T) {
	t.Run("invalid count", func(t *testing.T) {
		repo := &basketRepoStub{}
		for _, count := range []int{0, -1, -10} {
			err := AddProductUsecase(10, 5, count, repo)
			if !errors.Is(err, domain.ErrCountInvalid) {
				t.Fatalf("count %d: got err %v, want %v", count, err, domain.ErrCountInvalid)
			}
		}
	})

	t.Run("success", func(t *testing.T) {
		called := false
		repo := &basketRepoStub{
			addProduct: func(buyerID, productID int64, count int) error {
				called = true
				if buyerID != 10 || productID != 5 || count != 2 {
					t.Errorf("unexpected args: buyer=%d, prod=%d, count=%d", buyerID, productID, count)
				}
				return nil
			},
		}

		err := AddProductUsecase(10, 5, 2, repo)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("expected addProduct to be called")
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &basketRepoStub{
			addProduct: func(buyerID, productID int64, count int) error {
				return domain.ErrCountExceeded
			},
		}

		err := AddProductUsecase(10, 5, 100, repo)
		if !errors.Is(err, domain.ErrCountExceeded) {
			t.Fatalf("got err %v, want %v", err, domain.ErrCountExceeded)
		}
	})
}

func TestUpdateProductCountUsecase(t *testing.T) {
	t.Run("negative count", func(t *testing.T) {
		repo := &basketRepoStub{}
		err := UpdateProductCountUsecase(10, 5, -1, repo)
		if !errors.Is(err, domain.ErrCountInvalid) {
			t.Fatalf("got err %v, want %v", err, domain.ErrCountInvalid)
		}
	})

	t.Run("zero count removes item", func(t *testing.T) {
		called := false
		repo := &basketRepoStub{
			updateProductCount: func(buyerID, productID int64, count int) error {
				called = true
				if count != 0 {
					t.Errorf("expected count 0, got %d", count)
				}
				return nil
			},
		}

		err := UpdateProductCountUsecase(10, 5, 0, repo)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("expected updateProductCount to be called")
		}
	})

	t.Run("success update count", func(t *testing.T) {
		repo := &basketRepoStub{
			updateProductCount: func(buyerID, productID int64, count int) error {
				return nil
			},
		}

		err := UpdateProductCountUsecase(10, 5, 3, repo)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("not found in basket error", func(t *testing.T) {
		repo := &basketRepoStub{
			updateProductCount: func(buyerID, productID int64, count int) error {
				return domain.ErrProductNotFoundInBasket
			},
		}

		err := UpdateProductCountUsecase(10, 5, 3, repo)
		if !errors.Is(err, domain.ErrProductNotFoundInBasket) {
			t.Fatalf("got err %v, want %v", err, domain.ErrProductNotFoundInBasket)
		}
	})
}

func TestPreviewBasketUsecase(t *testing.T) {
	validPickup := domain.PickupPoint{
		UserID:    1,
		Longitude: 37.6,
		Latitude:  55.7,
	}

	t.Run("validation coordinates error", func(t *testing.T) {
		repo := &basketRepoStub{}
		_, valErrs, _, err := PreviewBasketUsecase(10, PreviewInDTO{
			Longitude: 200,
			Latitude:  -100,
		}, repo)

		if !errors.Is(err, domain.ErrInvalidCoordinates) {
			t.Fatalf("got err %v, want %v", err, domain.ErrInvalidCoordinates)
		}
		if valErrs.LongitudeErrMessage == "" || valErrs.LatitudeErrMessage == "" {
			t.Errorf("expected coordinate error messages: %+v", valErrs)
		}
	})

	t.Run("invalid promocode", func(t *testing.T) {
		repo := &basketRepoStub{
			getPromocode: func(code string) (*domain.Promocode, error) {
				return nil, domain.ErrPromocodeNotFoundOrExpired
			},
		}

		_, valErrs, _, err := PreviewBasketUsecase(10, PreviewInDTO{
			Longitude: 37.6,
			Latitude:  55.7,
			Promocode: "EXPIRED",
		}, repo)

		if !errors.Is(err, domain.ErrInvalidCoordinates) {
			t.Fatalf("got err %v, want %v", err, domain.ErrInvalidCoordinates)
		}
		if valErrs.PromocodeErrMessage == "" {
			t.Error("expected promocode error message")
		}
	})

	t.Run("empty basket", func(t *testing.T) {
		repo := &basketRepoStub{
			getBasket: func(buyerID int64) (domain.BasketSummary, error) {
				return domain.BasketSummary{Items: []domain.BasketItem{}}, nil
			},
		}

		_, _, _, err := PreviewBasketUsecase(10, PreviewInDTO{
			Longitude: 37.6,
			Latitude:  55.7,
		}, repo)

		if !errors.Is(err, domain.ErrBasketEmpty) {
			t.Fatalf("got err %v, want %v", err, domain.ErrBasketEmpty)
		}
	})

	t.Run("insufficient stock error", func(t *testing.T) {
		repo := &basketRepoStub{
			getBasket: func(buyerID int64) (domain.BasketSummary, error) {
				return domain.BasketSummary{
					Items: []domain.BasketItem{
						{
							ProductID:      1,
							Name:           "Item 1",
							SellerName:     "Seller A",
							Count:          5,
							AvailableCount: 3,
						},
					},
				}, nil
			},
		}

		_, _, stockErrs, err := PreviewBasketUsecase(10, PreviewInDTO{
			Longitude: 37.6,
			Latitude:  55.7,
		}, repo)

		if !errors.Is(err, domain.ErrCountExceeded) {
			t.Fatalf("got err %v, want %v", err, domain.ErrCountExceeded)
		}
		if len(stockErrs) != 1 || stockErrs[0].ProductID != 1 {
			t.Errorf("unexpected stock errors: %+v", stockErrs)
		}
	})

	t.Run("success without promocode", func(t *testing.T) {
		repo := &basketRepoStub{
			getBasket: func(buyerID int64) (domain.BasketSummary, error) {
				return domain.BasketSummary{
					Items: []domain.BasketItem{
						{
							ProductID:      1,
							Name:           "Shirt",
							Price:          1000,
							Discount:       20,
							Count:          2,
							AvailableCount: 10,
							CategoryID:     1,
						},
					},
				}, nil
			},
			getNearestPickupPoint: func(lat, lng float64) (domain.PickupPoint, error) {
				return validPickup, nil
			},
		}

		out, valErrs, stockErrs, err := PreviewBasketUsecase(10, PreviewInDTO{
			Longitude: 37.6,
			Latitude:  55.7,
		}, repo)

		if err != nil || valErrs.HasErrors() || len(stockErrs) > 0 {
			t.Fatalf("unexpected result: err=%v, valErrs=%+v, stockErrs=%+v", err, valErrs, stockErrs)
		}
		if out.Price != 2000 {
			t.Errorf("got Price %f, want 2000", out.Price)
		}
		if out.PriceWithDiscount != 1600 {
			t.Errorf("got PriceWithDiscount %f, want 1600", out.PriceWithDiscount)
		}
		if out.PriceResult != 1600 {
			t.Errorf("got PriceResult %f, want 1600", out.PriceResult)
		}
		if out.PickupPoint.Longitude != 37.6 || out.PickupPoint.Latitude != 55.7 {
			t.Errorf("unexpected pickup coords: %+v", out.PickupPoint)
		}
		expectedDate := time.Now().AddDate(0, 0, 3).Format("2006-01-02")
		if out.ExpectedDeliveryDate != expectedDate {
			t.Errorf("got delivery date %s, want %s", out.ExpectedDeliveryDate, expectedDate)
		}
	})

	t.Run("success with category promocode", func(t *testing.T) {
		cat1 := int64(1)
		repo := &basketRepoStub{
			getPromocode: func(code string) (*domain.Promocode, error) {
				return &domain.Promocode{
					ID:                  10,
					Code:                "CLOTHES10",
					AvailableCategoryID: &cat1,
					Discount:            10,
				}, nil
			},
			getBasket: func(buyerID int64) (domain.BasketSummary, error) {
				return domain.BasketSummary{
					Items: []domain.BasketItem{
						{
							ProductID:      1,
							Name:           "Shirt",
							Price:          1000,
							Discount:       20,
							Count:          1,
							AvailableCount: 10,
							CategoryID:     1, // Category 1 matches promocode
						},
						{
							ProductID:      2,
							Name:           "Book",
							Price:          500,
							Discount:       0,
							Count:          1,
							AvailableCount: 10,
							CategoryID:     2, // Category 2 does NOT match
						},
					},
				}, nil
			},
			getNearestPickupPoint: func(lat, lng float64) (domain.PickupPoint, error) {
				return validPickup, nil
			},
		}

		out, _, _, err := PreviewBasketUsecase(10, PreviewInDTO{
			Longitude: 37.6,
			Latitude:  55.7,
			Promocode: "CLOTHES10",
		}, repo)

		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		// Product 1: Price=1000, Discount=20% -> 800, Promo=10% -> factor = 0.8 * 0.9 = 0.72 -> 720
		// Product 2: Price=500, Discount=0%, Promo does not apply -> 500
		// Total result = 720 + 500 = 1220
		if out.PriceResult != 1220 {
			t.Errorf("got PriceResult %f, want 1220", out.PriceResult)
		}
	})
}
