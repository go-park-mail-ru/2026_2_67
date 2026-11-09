package orders_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/orders"
	"vibe_market/backend/internal/orders/domain"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type muxOrderRepoStub struct {
	listCalls, createCalls int
	ctx                    context.Context
	buyerID, pickupPointID int64
	promocode              string
}

func (s *muxOrderRepoStub) ListByBuyer(ctx context.Context, buyerID int64) ([]domain.OrderDetails, error) {
	s.listCalls++
	s.ctx, s.buyerID = ctx, buyerID
	return nil, nil
}

func (s *muxOrderRepoStub) CreateOrderTx(ctx context.Context, buyerID, pickupPointID int64, promocode string) (int64, []domain.ProductStockError, error) {
	s.createCalls++
	s.ctx, s.buyerID, s.pickupPointID, s.promocode = ctx, buyerID, pickupPointID, promocode
	return 123, nil, nil
}

func muxOrderToken(t *testing.T, secret []byte) string {
	t.Helper()
	claims := authDomain.AccessTokenPayload{}
	claims.UserID = 42
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour))
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + token
}

func TestNewServeMuxPreservesGetAndPost(t *testing.T) {
	// Exercise both registration orders at request time on the same mux.
	for _, methods := range [][]string{
		{http.MethodGet, http.MethodPost, http.MethodGet},
		{http.MethodPost, http.MethodGet, http.MethodPost},
	} {
		t.Run(strings.Join(methods, "-"), func(t *testing.T) {
			secret := []byte("orders-mux-test-secret")
			auth := muxOrderToken(t, secret)
			repo := &muxOrderRepoStub{}
			mux := orders.NewServeMux(repo, secret, zap.NewNop())
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "mux-context")
			listCalls, createCalls := 0, 0
			for _, method := range methods {
				r := httptest.NewRequest(method, "/?buyerId=99", strings.NewReader(`{"pickupPointId":17,"promocode":" SALE "}`)).WithContext(ctx)
				r.Header.Set("Authorization", auth)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, r)
				want := `{"orders":[]}`
				if method == http.MethodPost {
					createCalls++
					want = `{"orderId":123}`
				} else {
					listCalls++
				}
				if w.Code != http.StatusOK {
					t.Fatalf("%s status = %d, body = %s", method, w.Code, w.Body.String())
				}
				if got := w.Header().Get("Content-Type"); got != "application/json" {
					t.Errorf("%s Content-Type = %q", method, got)
				}
				var gotJSON, wantJSON any
				if err := json.Unmarshal(w.Body.Bytes(), &gotJSON); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal([]byte(want), &wantJSON); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(gotJSON, wantJSON) {
					t.Errorf("%s body = %s, want %s", method, w.Body.String(), want)
				}
				if repo.listCalls != listCalls || repo.createCalls != createCalls || repo.buyerID != 42 || repo.ctx != ctx {
					t.Errorf("%s unexpected repository calls: %+v", method, repo)
				}
				if method == http.MethodPost && (repo.pickupPointID != 17 || repo.promocode != "SALE") {
					t.Errorf("unexpected create arguments: %+v", repo)
				}
			}
		})
	}
}

func TestNewServeMuxUnknownSubpath(t *testing.T) {
	secret := []byte("orders-mux-test-secret")
	auth := muxOrderToken(t, secret)
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		for _, path := range []string{"/unknown", "/123", "/unknown/nested"} {
			t.Run(method+path, func(t *testing.T) {
				repo := &muxOrderRepoStub{}
				mux := orders.NewServeMux(repo, secret, zap.NewNop())
				r := httptest.NewRequest(method, path, strings.NewReader(`{"pickupPointId":17}`))
				r.Header.Set("Authorization", auth)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, r)
				if w.Code != http.StatusNotFound {
					t.Errorf("status = %d, want 404; body = %s", w.Code, w.Body.String())
				}
				if repo.listCalls != 0 || repo.createCalls != 0 {
					t.Errorf("unexpected repository calls: %+v", repo)
				}
			})
		}
	}
}

func TestNewServeMuxMethodNotAllowed(t *testing.T) {
	repo := &muxOrderRepoStub{}
	mux := orders.NewServeMux(repo, []byte("orders-mux-test-secret"), zap.NewNop())
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405; body = %s", w.Code, w.Body.String())
	}
	// Go's method-aware ServeMux may advertise HEAD implicitly alongside GET.
	allowed := map[string]bool{}
	for _, header := range w.Header().Values("Allow") {
		for _, method := range strings.Split(header, ",") {
			allowed[strings.TrimSpace(method)] = true
		}
	}
	if !allowed[http.MethodGet] || !allowed[http.MethodPost] {
		t.Errorf("Allow = %q, want GET and POST", w.Header().Values("Allow"))
	}
	for method := range allowed {
		if method != http.MethodGet && method != http.MethodHead && method != http.MethodPost {
			t.Errorf("unexpected allowed method %q", method)
		}
	}
	if repo.listCalls != 0 || repo.createCalls != 0 {
		t.Errorf("unexpected repository calls: %+v", repo)
	}
}
