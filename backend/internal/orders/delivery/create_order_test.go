package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/orders/domain"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var orderTestSecret = []byte("orders-test-secret")

type orderRepoStub struct {
	calls         int
	buyerID       int64
	pickupPointID int64
	promocode     string
	ctx           context.Context
	orderID       int64
	stockErrors   []domain.ProductStockError
	orders        []domain.OrderDetails
	err           error
}

func (s *orderRepoStub) CreateOrderTx(ctx context.Context, buyerID, pickupPointID int64, promoCode string) (int64, []domain.ProductStockError, error) {
	s.calls++
	s.ctx, s.buyerID, s.pickupPointID, s.promocode = ctx, buyerID, pickupPointID, promoCode
	return s.orderID, s.stockErrors, s.err
}

func (s *orderRepoStub) ListByBuyer(ctx context.Context, buyerID int64) ([]domain.OrderDetails, error) {
	s.calls++
	s.ctx, s.buyerID = ctx, buyerID
	return s.orders, s.err
}

func orderTestToken(t *testing.T, secret []byte, expiresAt time.Time) string {
	t.Helper()
	claims := authDomain.AccessTokenPayload{}
	claims.UserID = 42
	claims.ExpiresAt = jwt.NewNumericDate(expiresAt)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + token
}

func orderTestRequest(t *testing.T, repo *orderRepoStub, body, authorization string) (*httptest.ResponseRecorder, context.Context) {
	t.Helper()
	h := NewOrdersHandler(repo, orderTestSecret, zap.NewNop())
	r := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(body))
	if authorization != "" {
		r.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	h.CreateOrder(w, r)
	return w, r.Context()
}

func assertOrderJSON(t *testing.T, w *httptest.ResponseRecorder, status int, want string) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, status, w.Body.String())
	}
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	var gotJSON, wantJSON any
	if err := json.Unmarshal(w.Body.Bytes(), &gotJSON); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(want), &wantJSON); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotJSON, wantJSON) {
		t.Errorf("body = %s, want %s", w.Body.String(), want)
	}
}

func TestCreateOrderSuccess(t *testing.T) {
	auth := orderTestToken(t, orderTestSecret, time.Now().Add(time.Hour))
	for _, tc := range []struct {
		name, body, promocode string
	}{
		{"with promocode", `{"pickupPointId":17,"promocode":"  SALE  "}`, "SALE"},
		{"without promocode", `{"pickupPointId":17}`, ""},
		{"empty promocode", `{"pickupPointId":17,"promocode":""}`, ""},
		{"null promocode", `{"pickupPointId":17,"promocode":null}`, ""},
		{"unknown fields", `{"pickupPointId":17,"longitude":30,"latitude":60,"extra":true}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &orderRepoStub{orderID: 123}
			w, ctx := orderTestRequest(t, repo, tc.body, auth)
			assertOrderJSON(t, w, http.StatusOK, `{"orderId":123}`)
			if repo.calls != 1 || repo.buyerID != 42 || repo.pickupPointID != 17 || repo.promocode != tc.promocode || repo.ctx != ctx {
				t.Errorf("unexpected repository call: %+v", repo)
			}
		})
	}
}

func TestCreateOrderBadRequest(t *testing.T) {
	auth := orderTestToken(t, orderTestSecret, time.Now().Add(time.Hour))
	for _, tc := range []struct{ name, body string }{
		{"empty body", ""},
		{"invalid JSON", `{"pickupPointId":`},
		{"null body", `null`},
		{"array body", `[]`},
		{"missing ID", `{"promocode":"SALE"}`},
		{"null ID", `{"pickupPointId":null}`},
		{"old coordinates only", `{"longitude":30,"latitude":60}`},
		{"string ID", `{"pickupPointId":"17"}`},
		{"fractional ID", `{"pickupPointId":17.5}`},
		{"overflow ID", `{"pickupPointId":9223372036854775808}`},
		{"boolean ID", `{"pickupPointId":true}`},
		{"object ID", `{"pickupPointId":{}}`},
		{"array ID", `{"pickupPointId":[]}`},
		{"numeric promocode", `{"pickupPointId":17,"promocode":123}`},
		{"fractional promocode", `{"pickupPointId":17,"promocode":12.5}`},
		{"boolean promocode", `{"pickupPointId":17,"promocode":false}`},
		{"object promocode", `{"pickupPointId":17,"promocode":{}}`},
		{"array promocode", `{"pickupPointId":17,"promocode":[]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &orderRepoStub{}
			w, _ := orderTestRequest(t, repo, tc.body, auth)
			if w.Code != http.StatusBadRequest || w.Body.Len() != 0 || repo.calls != 0 {
				t.Fatalf("status = %d, body = %q, repository calls = %d", w.Code, w.Body.String(), repo.calls)
			}
		})
	}
}

func TestCreateOrderValidationErrors(t *testing.T) {
	auth := orderTestToken(t, orderTestSecret, time.Now().Add(time.Hour))
	for _, tc := range []struct {
		name, body, field string
		repoErr           error
		calls             int
	}{
		{"zero ID", `{"pickupPointId":0}`, "errPickupPointIdMessage", nil, 0},
		{"negative ID", `{"pickupPointId":-1}`, "errPickupPointIdMessage", nil, 0},
		{"unknown pickup point", `{"pickupPointId":17}`, "errPickupPointIdMessage", fmt.Errorf("lookup: %w", domain.ErrPickupPointInvalid), 1},
		{"invalid promocode", `{"pickupPointId":17,"promocode":"INVALID"}`, "promocodeErrMessage", fmt.Errorf("lookup: %w", domain.ErrPromocodeInvalid), 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &orderRepoStub{err: tc.repoErr}
			w, _ := orderTestRequest(t, repo, tc.body, auth)
			var response map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("invalid response: %v; body = %s", err, w.Body.String())
			}
			if response[tc.field] == "" {
				t.Fatalf("missing validation message for %s: %s", tc.field, w.Body.String())
			}
			want := map[string]string{
				"errMessage":              "Body содержит неверные данные",
				"errPickupPointIdMessage": "",
				"promocodeErrMessage":     "",
			}
			want[tc.field] = response[tc.field]
			wantJSON, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			assertOrderJSON(t, w, http.StatusUnprocessableEntity, string(wantJSON))
			if repo.calls != tc.calls {
				t.Errorf("repository calls = %d, want %d", repo.calls, tc.calls)
			}
		})
	}
}

func TestCreateOrderStockErrors(t *testing.T) {
	auth := orderTestToken(t, orderTestSecret, time.Now().Add(time.Hour))
	for _, tc := range []struct {
		name string
		repo orderRepoStub
		want string
	}{
		{
			name: "insufficient stock",
			repo: orderRepoStub{
				err: domain.ErrOrderStockInsufficient,
				stockErrors: []domain.ProductStockError{
					{ProductID: 7, Name: "Товар", Seller: "Продавец", ErrMessage: "недостаточно товара"},
					{ProductID: 8, Name: "Другой товар", Seller: "Магазин", ErrMessage: "товар закончился"},
				},
			},
			want: `{"errMessage":"при создании заказа произошла ошибка","productErrs":[{"productId":7,"name":"Товар","seller":"Продавец","errMessage":"недостаточно товара"},{"productId":8,"name":"Другой товар","seller":"Магазин","errMessage":"товар закончился"}]}`,
		},
		{
			name: "empty basket",
			repo: orderRepoStub{err: fmt.Errorf("create: %w", domain.ErrOrderBasketEmpty)},
			want: `{"errMessage":"при создании заказа произошла ошибка","productErrs":[]}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w, _ := orderTestRequest(t, &tc.repo, `{"pickupPointId":17}`, auth)
			assertOrderJSON(t, w, http.StatusUnprocessableEntity, tc.want)
			if tc.repo.calls != 1 {
				t.Errorf("repository calls = %d, want 1", tc.repo.calls)
			}
		})
	}
}

func TestCreateOrderInternalError(t *testing.T) {
	repo := &orderRepoStub{err: errors.New("database unavailable")}
	auth := orderTestToken(t, orderTestSecret, time.Now().Add(time.Hour))
	w, _ := orderTestRequest(t, repo, `{"pickupPointId":17}`, auth)
	if w.Code != http.StatusInternalServerError || w.Body.Len() != 0 || repo.calls != 1 {
		t.Fatalf("status = %d, body = %q, repository calls = %d", w.Code, w.Body.String(), repo.calls)
	}
}

func TestCreateOrderAuthentication(t *testing.T) {
	for _, tc := range []struct {
		name, auth string
		status     int
	}{
		{"missing header", "", http.StatusBadRequest},
		{"wrong scheme", "Basic token", http.StatusForbidden},
		{"malformed token", "Bearer invalid", http.StatusForbidden},
		{"wrong signature", orderTestToken(t, []byte("wrong-secret"), time.Now().Add(time.Hour)), http.StatusForbidden},
		{"expired token", orderTestToken(t, orderTestSecret, time.Now().Add(-time.Hour)), http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &orderRepoStub{}
			w, _ := orderTestRequest(t, repo, `{"pickupPointId":17}`, tc.auth)
			if w.Code != tc.status || w.Body.Len() != 0 || repo.calls != 0 {
				t.Fatalf("status = %d, want %d; body = %q, repository calls = %d", w.Code, tc.status, w.Body.String(), repo.calls)
			}
		})
	}
}
