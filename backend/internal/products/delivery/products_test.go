package delivery

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/products/domain"

	"go.uber.org/zap"
)

type productsRepoStub struct {
	params domain.ProductsSearchParams
	page   domain.ProductsPage
	err    error
	calls  int
}

func (s *productsRepoStub) Search(params domain.ProductsSearchParams) (domain.ProductsPage, error) {
	s.params = params
	s.calls++
	return s.page, s.err
}

func productsRequest(t *testing.T, repo *productsRepoStub, query string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/products"+query, nil)
	r = r.WithContext(context.WithValue(r.Context(), common.LoggerContextKey, zap.NewNop()))
	w := httptest.NewRecorder()
	NewProductsHandler(repo, zap.NewNop()).Products(w, r)
	if repo.calls != 1 {
		t.Fatalf("Search calls = %d, want 1", repo.calls)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	return w
}

func TestProductsFilters(t *testing.T) {
	createdAt := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	// Build the input independently of encodeCursor, so a matching encoder bug
	// cannot hide a decoding or mapping error.
	cursor := base64.RawURLEncoding.EncodeToString([]byte(`{"created_at":"2026-02-03T04:05:06Z","id":42,"rating":4.75}`))
	maxPrice := 250.5
	zero := 0.0
	negative := -1.0
	infinity := math.Inf(1)
	tests := []struct {
		name  string
		query string
		want  domain.ProductsSearchParams
	}{
		{name: "defaults"},
		{name: "explicit empty values", query: "?category=&query=&cursor=&limit=&priceMin=&priceMax=&sortBy="},
		{
			name: "all filters and repeated sellers",
			query: "?" + url.Values{
				"category": {"home & kitchen"}, "seller": {"seller one", "seller-two", "seller one"},
				"query": {"tea & coffee"}, "cursor": {cursor}, "limit": {"25"},
				"priceMin": {"12.25"}, "priceMax": {"250.5"}, "sortBy": {"popular"},
			}.Encode(),
			want: domain.ProductsSearchParams{
				Category: "home & kitchen", Seller: []string{"seller one", "seller-two", "seller one"},
				Query: "tea & coffee", Cursor: &domain.Cursor{CreatedAt: createdAt, ID: 42, Rating: 4.75},
				Limit: 25, PriceMin: 12.25, PriceMax: &maxPrice, SortBy: domain.SortByPopular,
			},
		},
		{name: "new sort", query: "?sortBy=new", want: domain.ProductsSearchParams{SortBy: domain.SortByNew}},
		{name: "unknown sort forwarded", query: "?sortBy=unknown", want: domain.ProductsSearchParams{SortBy: "unknown"}},
		{name: "zero values", query: "?limit=0&priceMin=0&priceMax=0", want: domain.ProductsSearchParams{PriceMax: &zero}},
		{name: "invalid numbers", query: "?limit=no&priceMin=no&priceMax=no"},
		{name: "negative limit", query: "?limit=-1"},
		{name: "fractional limit", query: "?limit=1.5"},
		{name: "overflow limit", query: "?limit=4294967296"},
		{name: "max uint32 limit", query: "?limit=4294967295", want: domain.ProductsSearchParams{Limit: 4294967295}},
		{name: "overflow prices", query: "?priceMin=1e999&priceMax=1e999"},
		{name: "negative prices forwarded", query: "?priceMin=-2&priceMax=-1", want: domain.ProductsSearchParams{PriceMin: -2, PriceMax: &negative}},
		{name: "infinite prices forwarded", query: "?priceMin=Inf&priceMax=Inf", want: domain.ProductsSearchParams{PriceMin: infinity, PriceMax: &infinity}},
		{name: "invalid base64 cursor", query: "?cursor=%25%25%25", want: domain.ProductsSearchParams{}},
		{name: "invalid JSON cursor", query: "?cursor=" + base64.RawURLEncoding.EncodeToString([]byte("not JSON"))},
		{name: "missing cursor date", query: "?cursor=" + base64.RawURLEncoding.EncodeToString([]byte(`{"id":42,"rating":4.75}`))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &productsRepoStub{}
			productsRequest(t, repo, tt.query)
			if !reflect.DeepEqual(repo.params, tt.want) {
				t.Errorf("Search params = %#v, want %#v", repo.params, tt.want)
			}
		})
	}
}

func TestProductsJSONSchema(t *testing.T) {
	product := domain.Product{
		ID: 7, Name: "Tea", Price: 200, Discount: 25,
		CategoryID: 10, SellerID: 20, AvailableCount: 99,
	}
	tests := []struct {
		name string
		page domain.ProductsPage
		want string
	}{
		{
			name: "populated",
			page: domain.ProductsPage{
				Products: []domain.Product{product}, PictureURLs: [][]string{{"/tea.png", "/tea-2.png"}},
				Ratings: []float64{4.5}, ReviewCounts: []uint{12},
			},
			want: `{"products":[{"id":7,"name":"Tea","pictureUrls":["/tea.png","/tea-2.png"],"price":200,"discount":25,"priceResult":150,"rating":4.5,"reviewsCount":12}],"nextCursor":""}`,
		},
		{name: "nil products", want: `{"products":[],"nextCursor":""}`},
		{name: "empty products", page: domain.ProductsPage{Products: []domain.Product{}}, want: `{"products":[],"nextCursor":""}`},
		{
			name: "missing metadata",
			page: domain.ProductsPage{Products: []domain.Product{product}},
			want: `{"products":[{"id":7,"name":"Tea","pictureUrls":[],"price":200,"discount":25,"priceResult":150,"rating":0,"reviewsCount":0}],"nextCursor":""}`,
		},
		{
			name: "short metadata and nil pictures",
			page: domain.ProductsPage{
				Products:    []domain.Product{product, {ID: 8, Name: "Cup", Price: 50}},
				PictureURLs: [][]string{nil}, Ratings: []float64{3.5}, ReviewCounts: []uint{2},
			},
			want: `{"products":[{"id":7,"name":"Tea","pictureUrls":[],"price":200,"discount":25,"priceResult":150,"rating":3.5,"reviewsCount":2},{"id":8,"name":"Cup","pictureUrls":[],"price":50,"discount":0,"priceResult":50,"rating":0,"reviewsCount":0}],"nextCursor":""}`,
		},
		{
			name: "empty picture list",
			page: domain.ProductsPage{Products: []domain.Product{product}, PictureURLs: [][]string{{}}},
			want: `{"products":[{"id":7,"name":"Tea","pictureUrls":[],"price":200,"discount":25,"priceResult":150,"rating":0,"reviewsCount":0}],"nextCursor":""}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := productsRequest(t, &productsRepoStub{page: tt.page}, "")
			assertProductsJSON(t, w.Body.String(), tt.want)
		})
	}
}

func assertProductsJSON(t *testing.T, got, want string) {
	t.Helper()
	var actual, expected any
	if err := json.Unmarshal([]byte(got), &actual); err != nil {
		t.Fatalf("invalid response JSON: %v; body: %s", err, got)
	}
	if err := json.Unmarshal([]byte(want), &expected); err != nil {
		t.Fatalf("invalid expected JSON: %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("JSON = %s, want %s", got, want)
	}
}

func TestProductsNextCursor(t *testing.T) {
	createdAt := time.Date(2026, 2, 3, 7, 5, 6, 123, time.FixedZone("offset", 3*60*60))
	repo := &productsRepoStub{page: domain.ProductsPage{
		Products:   []domain.Product{{ID: 42, Name: "Tea", Price: 10}},
		NextCursor: &domain.Cursor{CreatedAt: createdAt, ID: 42, Rating: 4.75},
	}}
	w := productsRequest(t, repo, "")
	var response struct {
		Products   []json.RawMessage `json:"products"`
		NextCursor string            `json:"nextCursor"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Products) != 1 || response.NextCursor == "" {
		t.Fatalf("unexpected response: %s", w.Body.String())
	}
	raw, err := base64.RawURLEncoding.DecodeString(response.NextCursor)
	if err != nil {
		t.Fatalf("nextCursor is not raw URL base64: %v", err)
	}
	assertProductsJSON(t, string(raw), `{"created_at":"2026-02-03T04:05:06.000000123Z","id":42,"rating":4.75}`)
	// The returned token is usable as the next request's cursor.
	nextRepo := &productsRepoStub{}
	productsRequest(t, nextRepo, "?cursor="+url.QueryEscape(response.NextCursor))
	if nextRepo.params.Cursor == nil || nextRepo.params.Cursor.ID != 42 ||
		nextRepo.params.Cursor.Rating != 4.75 || !nextRepo.params.Cursor.CreatedAt.Equal(createdAt) {
		t.Fatalf("next request cursor = %#v", nextRepo.params.Cursor)
	}
}

func TestProductsRepositoryErrorsReturnEmptyOK(t *testing.T) {
	for _, repoErr := range []error{
		errors.New("database unavailable"), domain.ErrInvalidSortBy,
		domain.ErrInvalidPriceRange, domain.ErrInvalidCursor,
		fmt.Errorf("search: %w", domain.ErrInvalidSortBy),
		fmt.Errorf("search: %w", domain.ErrInvalidPriceRange),
		fmt.Errorf("search: %w", domain.ErrInvalidCursor),
	} {
		t.Run(repoErr.Error(), func(t *testing.T) {
			repo := &productsRepoStub{
				err: repoErr,
				page: domain.ProductsPage{
					Products:   []domain.Product{{ID: 1, Name: "discard me"}},
					NextCursor: &domain.Cursor{CreatedAt: time.Now(), ID: 1},
				},
			}
			w := productsRequest(t, repo, "")
			assertProductsJSON(t, w.Body.String(), `{"products":[],"nextCursor":""}`)
		})
	}
}

func TestDecodeCursor(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "invalid JSON", json: "{"},
		{name: "null", json: "null"},
		{name: "array", json: "[]"},
		{name: "missing fields", json: "{}"},
		{name: "missing date", json: `{"id":1,"rating":4}`},
		{name: "zero date", json: `{"created_at":"0001-01-01T00:00:00Z","id":1}`},
		{name: "invalid date", json: `{"created_at":"yesterday","id":1}`},
		{name: "invalid id type", json: `{"created_at":"2026-02-03T04:05:06Z","id":"1"}`},
		{name: "invalid rating type", json: `{"created_at":"2026-02-03T04:05:06Z","rating":"high"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeCursor(base64.RawURLEncoding.EncodeToString([]byte(tt.json)))
			if !errors.Is(err, ErrInvalidCursorBase64) || got != (cursorBase64{}) {
				t.Errorf("decodeCursor = %#v, %v; want zero cursor, ErrInvalidCursorBase64", got, err)
			}
		})
	}
	for _, input := range []string{"", "%", "e30="} {
		t.Run("invalid base64 "+input, func(t *testing.T) {
			if _, err := decodeCursor(input); !errors.Is(err, ErrInvalidCursorBase64) {
				t.Errorf("decodeCursor(%q) error = %v", input, err)
			}
		})
	}
	t.Run("date alone is accepted by current helper", func(t *testing.T) {
		got, err := decodeCursor(base64.RawURLEncoding.EncodeToString([]byte(`{"created_at":"2026-02-03T04:05:06Z"}`)))
		if err != nil || got.CreatedAt.IsZero() || got.ID != 0 || got.Rating != 0 {
			t.Errorf("decodeCursor = %#v, %v", got, err)
		}
	})
}

func TestEncodeCursor(t *testing.T) {
	createdAt := time.Date(2026, 2, 3, 7, 5, 6, 0, time.FixedZone("offset", 3*60*60))
	token, err := encodeCursor(createdAt, 42, 4.75)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(token, "+/=") {
		t.Fatalf("cursor is not unpadded URL-safe base64: %q", token)
	}
	got, err := decodeCursor(token)
	if err != nil || !got.CreatedAt.Equal(createdAt) || got.CreatedAt.Location() != time.UTC || got.ID != 42 || got.Rating != 4.75 {
		t.Fatalf("round trip = %#v, %v", got, err)
	}
	for _, rating := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		t.Run(fmt.Sprint(rating), func(t *testing.T) {
			if token, err := encodeCursor(createdAt, 1, rating); err == nil || token != "" {
				t.Errorf("encodeCursor = %q, %v; want empty token and error", token, err)
			}
		})
	}
	if token, err := encodeCursor(time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC), 1, 0); err == nil || token != "" {
		t.Errorf("out-of-range date: encodeCursor = %q, %v", token, err)
	}
}

func TestProductsInvalidNextCursor(t *testing.T) {
	repo := &productsRepoStub{page: domain.ProductsPage{
		NextCursor: &domain.Cursor{CreatedAt: time.Now(), ID: 1, Rating: math.NaN()},
	}}
	w := productsRequest(t, repo, "")
	assertProductsJSON(t, w.Body.String(), `{"products":[],"nextCursor":""}`)
}

func TestProductsFallbackLogger(t *testing.T) {
	repo := &productsRepoStub{}
	r := httptest.NewRequest(http.MethodGet, "/products?limit=invalid", nil)
	// A typed nil exercises the fallback without the unchecked context type
	// assertion panicking on an absent context value.
	r = r.WithContext(context.WithValue(r.Context(), common.LoggerContextKey, (*zap.Logger)(nil)))
	w := httptest.NewRecorder()
	NewProductsHandler(repo, zap.NewNop()).Products(w, r)
	if w.Code != http.StatusOK || repo.calls != 1 {
		t.Fatalf("status = %d, Search calls = %d", w.Code, repo.calls)
	}
	assertProductsJSON(t, w.Body.String(), `{"products":[],"nextCursor":""}`)
}

type productsFailingWriter struct {
	header http.Header
}

func (w *productsFailingWriter) Header() http.Header { return w.header }
func (w *productsFailingWriter) WriteHeader(int)     {}
func (w *productsFailingWriter) Write([]byte) (int, error) {
	return 0, errors.New("connection closed")
}

func TestProductsResponseWriteError(t *testing.T) {
	repo := &productsRepoStub{}
	r := httptest.NewRequest(http.MethodGet, "/products", nil)
	r = r.WithContext(context.WithValue(r.Context(), common.LoggerContextKey, zap.NewNop()))
	w := &productsFailingWriter{header: make(http.Header)}
	NewProductsHandler(repo, zap.NewNop()).Products(w, r)
	if repo.calls != 1 || w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Search calls = %d, headers = %v", repo.calls, w.Header())
	}
}
