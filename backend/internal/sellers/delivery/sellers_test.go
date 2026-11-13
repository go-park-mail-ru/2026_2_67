package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
	"vibe_market/backend/internal/sellers/domain"

	"go.uber.org/zap"
)

type sellerRepoStub struct {
	params  domain.SellerSearchParams
	sellers []domain.Seller
	err     error
	calls   int
}

func (s *sellerRepoStub) Search(params domain.SellerSearchParams) ([]domain.Seller, error) {
	s.params = params
	s.calls++
	return s.sellers, s.err
}

func TestSellersFilters(t *testing.T) {
	maxPrice := 99.75
	zero := 0.0
	for _, tt := range []struct {
		name  string
		query string
		want  domain.SellerSearchParams
	}{
		{
			name:  "all filters",
			query: "?limit=25&offset=10&category=home+goods&query=Alice+%26+Bob&priceMin=12.5&priceMax=99.75",
			want: domain.SellerSearchParams{
				Limit: 25, Offset: 10, Category: "home goods", Query: "Alice & Bob",
				PriceMin: 12.5, PriceMax: &maxPrice,
			},
		},
		{name: "defaults"},
		{name: "empty values", query: "?limit=&offset=&category=&query=&priceMin=&priceMax="},
		{name: "invalid numbers", query: "?limit=no&offset=no&priceMin=no&priceMax=no&category=books&query=test", want: domain.SellerSearchParams{Category: "books", Query: "test"}},
		{name: "negative unsigned numbers", query: "?limit=-1&offset=-2"},
		{name: "fractional unsigned numbers", query: "?limit=1.5&offset=2.5"},
		{name: "overflow", query: "?limit=4294967296&offset=4294967296&priceMin=1e999&priceMax=1e999"},
		{name: "explicit zero", query: "?limit=0&offset=0&priceMin=0&priceMax=0", want: domain.SellerSearchParams{PriceMax: &zero}},
		{name: "invalid minimum preserves maximum", query: "?priceMin=no&priceMax=99.75", want: domain.SellerSearchParams{PriceMax: &maxPrice}},
		{name: "invalid maximum preserves minimum", query: "?priceMin=12.5&priceMax=no", want: domain.SellerSearchParams{PriceMin: 12.5}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &sellerRepoStub{}
			w := httptest.NewRecorder()
			NewSellersHandler(repo, zap.NewNop()).Sellers(w, httptest.NewRequest(http.MethodGet, "/sellers"+tt.query, nil))
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
			}
			if repo.calls != 1 {
				t.Fatalf("Search calls = %d, want 1", repo.calls)
			}
			if !reflect.DeepEqual(repo.params, tt.want) {
				t.Errorf("Search params = %+v, want %+v", repo.params, tt.want)
			}
		})
	}
}

func TestSellersResponse(t *testing.T) {
	description, avatar := "private description", "avatar.png"
	for _, tt := range []struct {
		name    string
		sellers []domain.Seller
		want    string
	}{
		{
			name: "only id and name",
			sellers: []domain.Seller{
				{UserID: 42, Name: "Alice", Description: &description, AvatarURL: &avatar, CreatedAt: time.Unix(1, 0), UpdatedAt: time.Unix(2, 0)},
				{UserID: 7, Name: "Bob"},
			},
			want: `{"sellers":[{"id":42,"name":"Alice"},{"id":7,"name":"Bob"}]}`,
		},
		{name: "nil result", want: `{"sellers":[]}`},
		{name: "empty result", sellers: []domain.Seller{}, want: `{"sellers":[]}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &sellerRepoStub{sellers: tt.sellers}
			w := httptest.NewRecorder()
			NewSellersHandler(repo, zap.NewNop()).Sellers(w, httptest.NewRequest(http.MethodGet, "/sellers", nil))
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
			}
			if got := w.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			var got, want any
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatalf("decode expected JSON: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("response = %s, want %s", w.Body.String(), tt.want)
			}
		})
	}
}

func TestSellersRepoError(t *testing.T) {
	repo := &sellerRepoStub{
		sellers: []domain.Seller{{UserID: 42, Name: "Alice"}},
		err:     errors.New("private repository failure"),
	}
	w := httptest.NewRecorder()
	NewSellersHandler(repo, zap.NewNop()).Sellers(w, httptest.NewRequest(http.MethodGet, "/sellers", nil))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if repo.calls != 1 {
		t.Errorf("Search calls = %d, want 1", repo.calls)
	}
	if w.Body.Len() != 0 {
		t.Errorf("error response must be empty, got %q", w.Body.String())
	}
}
