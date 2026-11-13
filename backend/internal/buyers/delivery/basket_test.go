package delivery

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/buyers/domain"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var testJWTSecret = []byte("basket-test-jwt-secret")

type basketDeliveryRepoStub struct {
	getBasket             func(buyerID int64) (domain.BasketSummary, error)
	addProduct            func(buyerID int64, productID int64, count int) error
	updateProductCount    func(buyerID int64, productID int64, count int) error
	getNearestPickupPoint func(lat, lng float64) (domain.PickupPoint, error)
	getPromocode          func(code string) (*domain.Promocode, error)
}

var _ domain.BasketRepo = (*basketDeliveryRepoStub)(nil)

func (s *basketDeliveryRepoStub) GetBasket(buyerID int64) (domain.BasketSummary, error) {
	if s.getBasket != nil {
		return s.getBasket(buyerID)
	}
	panic("unexpected GetBasket call")
}

func (s *basketDeliveryRepoStub) AddProduct(buyerID int64, productID int64, count int) error {
	if s.addProduct != nil {
		return s.addProduct(buyerID, productID, count)
	}
	panic("unexpected AddProduct call")
}

func (s *basketDeliveryRepoStub) UpdateProductCount(buyerID int64, productID int64, count int) error {
	if s.updateProductCount != nil {
		return s.updateProductCount(buyerID, productID, count)
	}
	panic("unexpected UpdateProductCount call")
}

func (s *basketDeliveryRepoStub) GetNearestPickupPoint(lat, lng float64) (domain.PickupPoint, error) {
	if s.getNearestPickupPoint != nil {
		return s.getNearestPickupPoint(lat, lng)
	}
	panic("unexpected GetNearestPickupPoint call")
}

func (s *basketDeliveryRepoStub) GetPromocode(code string) (*domain.Promocode, error) {
	if s.getPromocode != nil {
		return s.getPromocode(code)
	}
	panic("unexpected GetPromocode call")
}

func createTestToken(t *testing.T, userID int64, secret []byte, expiresAt time.Time) string {
	t.Helper()
	claims := authDomain.AccessTokenPayload{
		User: authDomain.User{UserID: userID},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + token
}

func TestGetBasketDelivery(t *testing.T) {
	token := createTestToken(t, 10, testJWTSecret, time.Now().Add(time.Hour))

	t.Run("success", func(t *testing.T) {
		repo := &basketDeliveryRepoStub{
			getBasket: func(buyerID int64) (domain.BasketSummary, error) {
				return domain.BasketSummary{
					Items: []domain.BasketItem{
						{ProductID: 1, Name: "Book", Price: 100, Count: 2, PriceResult: 100},
					},
					BasketSum:       200,
					BasketSumResult: 200,
				}, nil
			},
		}

		handler := NewBuyersHandler(nil, repo, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodGet, "/api/v1/buyers/10/basket", nil)
		r.SetPathValue("id", "10")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.GetBasket(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("invalid id in path", func(t *testing.T) {
		handler := NewBuyersHandler(nil, nil, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodGet, "/api/v1/buyers/abc/basket", nil)
		r.SetPathValue("id", "abc")
		w := httptest.NewRecorder()

		handler.GetBasket(w, r)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized user mismatch", func(t *testing.T) {
		handler := NewBuyersHandler(nil, nil, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodGet, "/api/v1/buyers/99/basket", nil)
		r.SetPathValue("id", "99")
		r.Header.Set("Authorization", token) // token for id=10
		w := httptest.NewRecorder()

		handler.GetBasket(w, r)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("buyer not found", func(t *testing.T) {
		repo := &basketDeliveryRepoStub{
			getBasket: func(buyerID int64) (domain.BasketSummary, error) {
				return domain.BasketSummary{}, domain.ErrBuyerNotFound
			},
		}

		handler := NewBuyersHandler(nil, repo, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodGet, "/api/v1/buyers/10/basket", nil)
		r.SetPathValue("id", "10")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.GetBasket(w, r)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		repo := &basketDeliveryRepoStub{
			getBasket: func(buyerID int64) (domain.BasketSummary, error) {
				return domain.BasketSummary{}, errors.New("db down")
			},
		}

		handler := NewBuyersHandler(nil, repo, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodGet, "/api/v1/buyers/10/basket", nil)
		r.SetPathValue("id", "10")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.GetBasket(w, r)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestAddBasketProductDelivery(t *testing.T) {
	token := createTestToken(t, 10, testJWTSecret, time.Now().Add(time.Hour))

	t.Run("success", func(t *testing.T) {
		repo := &basketDeliveryRepoStub{
			addProduct: func(buyerID, productID int64, count int) error {
				return nil
			},
		}

		handler := NewBuyersHandler(nil, repo, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodPost, "/api/v1/buyers/10/basket/products", bytes.NewBufferString(`{"productId":1,"count":2}`))
		r.SetPathValue("id", "10")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.AddBasketProduct(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("invalid json body", func(t *testing.T) {
		handler := NewBuyersHandler(nil, nil, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodPost, "/api/v1/buyers/10/basket/products", bytes.NewBufferString(`{invalid`))
		r.SetPathValue("id", "10")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.AddBasketProduct(w, r)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		handler := NewBuyersHandler(nil, nil, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodPost, "/api/v1/buyers/10/basket/products", bytes.NewBufferString(`{"productId":1}`))
		r.SetPathValue("id", "10")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.AddBasketProduct(w, r)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("product not found returns 404", func(t *testing.T) {
		repo := &basketDeliveryRepoStub{
			addProduct: func(buyerID, productID int64, count int) error {
				return domain.ErrProductNotFound
			},
		}

		handler := NewBuyersHandler(nil, repo, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodPost, "/api/v1/buyers/10/basket/products", bytes.NewBufferString(`{"productId":999,"count":1}`))
		r.SetPathValue("id", "10")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.AddBasketProduct(w, r)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("count exceeded returns 422", func(t *testing.T) {
		repo := &basketDeliveryRepoStub{
			addProduct: func(buyerID, productID int64, count int) error {
				return domain.ErrCountExceeded
			},
		}

		handler := NewBuyersHandler(nil, repo, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodPost, "/api/v1/buyers/10/basket/products", bytes.NewBufferString(`{"productId":1,"count":100}`))
		r.SetPathValue("id", "10")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.AddBasketProduct(w, r)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
	})
}

func TestUpdateBasketProductDelivery(t *testing.T) {
	token := createTestToken(t, 10, testJWTSecret, time.Now().Add(time.Hour))

	t.Run("success", func(t *testing.T) {
		repo := &basketDeliveryRepoStub{
			updateProductCount: func(buyerID, productID int64, count int) error {
				return nil
			},
		}

		handler := NewBuyersHandler(nil, repo, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodPatch, "/api/v1/buyers/10/basket/products/5", bytes.NewBufferString(`{"count":3}`))
		r.SetPathValue("buyer_id", "10")
		r.SetPathValue("product_id", "5")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.UpdateBasketProduct(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("not found in basket returns 404", func(t *testing.T) {
		repo := &basketDeliveryRepoStub{
			updateProductCount: func(buyerID, productID int64, count int) error {
				return domain.ErrProductNotFoundInBasket
			},
		}

		handler := NewBuyersHandler(nil, repo, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodPatch, "/api/v1/buyers/10/basket/products/5", bytes.NewBufferString(`{"count":3}`))
		r.SetPathValue("buyer_id", "10")
		r.SetPathValue("product_id", "5")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.UpdateBasketProduct(w, r)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("invalid count returns 422", func(t *testing.T) {
		repo := &basketDeliveryRepoStub{
			updateProductCount: func(buyerID, productID int64, count int) error {
				return domain.ErrCountInvalid
			},
		}

		handler := NewBuyersHandler(nil, repo, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodPatch, "/api/v1/buyers/10/basket/products/5", bytes.NewBufferString(`{"count":-1}`))
		r.SetPathValue("buyer_id", "10")
		r.SetPathValue("product_id", "5")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.UpdateBasketProduct(w, r)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
	})
}

func TestPreviewBasketDelivery(t *testing.T) {
	token := createTestToken(t, 10, testJWTSecret, time.Now().Add(time.Hour))

	t.Run("success", func(t *testing.T) {
		repo := &basketDeliveryRepoStub{
			getBasket: func(buyerID int64) (domain.BasketSummary, error) {
				return domain.BasketSummary{
					Items: []domain.BasketItem{
						{ProductID: 1, Name: "Item", Price: 100, Count: 1, AvailableCount: 5},
					},
				}, nil
			},
			getNearestPickupPoint: func(lat, lng float64) (domain.PickupPoint, error) {
				return domain.PickupPoint{UserID: 1, Longitude: 37.6, Latitude: 55.7}, nil
			},
		}

		handler := NewBuyersHandler(nil, repo, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodPost, "/api/v1/buyers/10/basket/preview", bytes.NewBufferString(`{"longitude":37.6,"latitude":55.7}`))
		r.SetPathValue("id", "10")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.PreviewBasket(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("coordinate validation error returns 422", func(t *testing.T) {
		handler := NewBuyersHandler(nil, &basketDeliveryRepoStub{}, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodPost, "/api/v1/buyers/10/basket/preview", bytes.NewBufferString(`{"longitude":300.0,"latitude":55.7}`))
		r.SetPathValue("id", "10")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.PreviewBasket(w, r)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
		var resp previewValidationResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.LongitudeErrMessage == "" {
			t.Error("expected LongitudeErrMessage")
		}
	})

	t.Run("stock error returns 422 with productsErr", func(t *testing.T) {
		repo := &basketDeliveryRepoStub{
			getBasket: func(buyerID int64) (domain.BasketSummary, error) {
				return domain.BasketSummary{
					Items: []domain.BasketItem{
						{ProductID: 1, Name: "Item", Price: 100, Count: 10, AvailableCount: 2},
					},
				}, nil
			},
		}

		handler := NewBuyersHandler(nil, repo, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodPost, "/api/v1/buyers/10/basket/preview", bytes.NewBufferString(`{"longitude":37.6,"latitude":55.7}`))
		r.SetPathValue("id", "10")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.PreviewBasket(w, r)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
		var resp previewStockErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if len(resp.ProductsErr) != 1 {
			t.Fatalf("expected 1 product err, got %d", len(resp.ProductsErr))
		}
	})

	t.Run("empty basket returns 422", func(t *testing.T) {
		repo := &basketDeliveryRepoStub{
			getBasket: func(buyerID int64) (domain.BasketSummary, error) {
				return domain.BasketSummary{Items: []domain.BasketItem{}}, nil
			},
		}

		handler := NewBuyersHandler(nil, repo, nil, testJWTSecret, zap.NewNop(), nil)
		r := httptest.NewRequest(http.MethodPost, "/api/v1/buyers/10/basket/preview", bytes.NewBufferString(`{"longitude":37.6,"latitude":55.7}`))
		r.SetPathValue("id", "10")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		handler.PreviewBasket(w, r)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
	})
}
