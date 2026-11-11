package usecase

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"vibe_market/backend/internal/products/domain"
)

type productRepoStub struct {
	calls  int
	params domain.ProductsSearchParams
	page   domain.ProductsPage
	err    error
}

func (s *productRepoStub) Search(params domain.ProductsSearchParams) (domain.ProductsPage, error) {
	s.calls++
	s.params = params
	return s.page, s.err
}

var _ domain.ProductRepo = (*productRepoStub)(nil)

func TestProductsPassesSearchParams(t *testing.T) {
	maxPrice := 250.75
	cursor := &domain.Cursor{
		CreatedAt: time.Date(2026, time.October, 27, 12, 30, 0, 0, time.UTC),
		ID:        42,
		Rating:    4.75,
	}
	tests := []struct {
		name   string
		params domain.ProductsSearchParams
	}{
		{
			name: "all fields",
			params: domain.ProductsSearchParams{
				Category: " Books ", Seller: []string{"Seller B", "Seller A", "Seller B"},
				Query: `50%_\`, Cursor: cursor, Limit: 17, PriceMin: 12.25,
				PriceMax: &maxPrice, SortBy: domain.SortByPopular,
			},
		},
		{name: "zero values"},
		{
			name:   "empty sellers and new sort",
			params: domain.ProductsSearchParams{Seller: []string{}, SortBy: domain.SortByNew},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &productRepoStub{}
			if _, err := Products(tt.params, repo); err != nil {
				t.Fatalf("Products() error = %v", err)
			}
			if repo.calls != 1 {
				t.Fatalf("Search calls = %d, want 1", repo.calls)
			}
			if !reflect.DeepEqual(repo.params, tt.params) {
				t.Errorf("Search params = %#v, want %#v", repo.params, tt.params)
			}
			if repo.params.Cursor != tt.params.Cursor || repo.params.PriceMax != tt.params.PriceMax {
				t.Error("Search must receive the original cursor and price pointers")
			}
		})
	}
}

func TestProductsPricesAndPagePassthrough(t *testing.T) {
	createdAt := time.Date(2026, time.October, 27, 12, 0, 0, 0, time.UTC)
	description := "product description"
	page := domain.ProductsPage{
		Products: []domain.Product{
			{
				ID: 9, CategoryID: 3, SellerID: 7, Price: 125.5, Discount: 0,
				Name: "zero discount", Description: &description, AvailableCount: 11,
				CreatedAt: createdAt, UpdatedAt: createdAt.Add(time.Hour),
			},
			{ID: 2, Price: 200, Discount: 25, Name: "partial discount"},
			{ID: 8, Price: 99.99, Discount: 100, Name: "full discount"},
			{ID: 1, Price: 19.99, Discount: 33, Name: "fractional result"},
		},
		PictureURLs:  [][]string{{"nine-1.jpg", "nine-2.jpg"}, nil, {}, {"one.jpg"}},
		Ratings:      []float64{4.75, 0, 3.25, 5},
		ReviewCounts: []uint{12, 0, 3, 42},
		NextCursor:   &domain.Cursor{CreatedAt: createdAt, ID: 1, Rating: 5},
	}
	repo := &productRepoStub{page: page}
	got, err := Products(domain.ProductsSearchParams{}, repo)
	if err != nil {
		t.Fatalf("Products() error = %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("Search calls = %d, want 1", repo.calls)
	}
	if !reflect.DeepEqual(got.Products, page.Products) {
		t.Errorf("Products = %#v, want unchanged products in repository order %#v", got.Products, page.Products)
	}
	wantPrices := []float64{125.5, 150, 0, 13.3933}
	if len(got.PricesResult) != len(wantPrices) {
		t.Fatalf("PricesResult length = %d, want %d", len(got.PricesResult), len(wantPrices))
	}
	for i, want := range wantPrices {
		if actual := got.PricesResult[i]; math.IsNaN(actual) || math.IsInf(actual, 0) || math.Abs(actual-want) > 1e-9 {
			t.Errorf("PricesResult[%d] = %.15g, want %.15g (tolerance 1e-9)", i, actual, want)
		}
	}
	if !reflect.DeepEqual(got.PictureURLs, page.PictureURLs) {
		t.Errorf("PictureURLs = %#v, want %#v", got.PictureURLs, page.PictureURLs)
	}
	if !reflect.DeepEqual(got.Ratings, page.Ratings) {
		t.Errorf("Ratings = %v, want %v", got.Ratings, page.Ratings)
	}
	if !reflect.DeepEqual(got.ReviewCounts, page.ReviewCounts) {
		t.Errorf("ReviewCounts = %v, want %v", got.ReviewCounts, page.ReviewCounts)
	}
	if got.NextCursor != page.NextCursor {
		t.Errorf("NextCursor = %p, want original pointer %p", got.NextCursor, page.NextCursor)
	}
}

func TestProductsEmptyPages(t *testing.T) {
	tests := []struct {
		name string
		page domain.ProductsPage
	}{
		{name: "nil slices"},
		{
			name: "empty slices",
			page: domain.ProductsPage{
				Products: []domain.Product{}, PictureURLs: [][]string{},
				Ratings: []float64{}, ReviewCounts: []uint{},
			},
		},
		{
			name: "metadata without products",
			page: domain.ProductsPage{
				PictureURLs: [][]string{nil, {}, {"image.jpg"}},
				Ratings:     []float64{4.5}, ReviewCounts: []uint{2},
				NextCursor: &domain.Cursor{ID: 7, Rating: 4.5},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &productRepoStub{page: tt.page}
			got, err := Products(domain.ProductsSearchParams{}, repo)
			if err != nil {
				t.Fatalf("Products() error = %v", err)
			}
			if repo.calls != 1 {
				t.Fatalf("Search calls = %d, want 1", repo.calls)
			}
			want := ProductsOutDTO{
				Products: tt.page.Products, PricesResult: []float64{},
				PictureURLs: tt.page.PictureURLs, Ratings: tt.page.Ratings,
				ReviewCounts: tt.page.ReviewCounts, NextCursor: tt.page.NextCursor,
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Products() = %#v, want %#v", got, want)
			}
			if got.NextCursor != tt.page.NextCursor {
				t.Error("NextCursor must preserve the repository pointer")
			}
		})
	}
}

func TestProductsRepositoryErrorsDiscardPartialData(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "invalid sort", err: domain.ErrInvalidSortBy},
		{name: "invalid price range", err: domain.ErrInvalidPriceRange},
		{name: "invalid cursor", err: domain.ErrInvalidCursor},
		{name: "unexpected error", err: errors.New("repository unavailable")},
	}
	for _, tt := range tests {
		for _, wrapped := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/wrapped=%t", tt.name, wrapped), func(t *testing.T) {
				repoErr := tt.err
				if wrapped {
					repoErr = fmt.Errorf("search products: %w", tt.err)
				}
				repo := &productRepoStub{
					err: repoErr,
					page: domain.ProductsPage{
						Products:    []domain.Product{{ID: 1, Price: 100, Discount: 25}},
						PictureURLs: [][]string{{"partial.jpg"}}, Ratings: []float64{4},
						ReviewCounts: []uint{3}, NextCursor: &domain.Cursor{ID: 1},
					},
				}
				got, err := Products(domain.ProductsSearchParams{}, repo)
				if repo.calls != 1 {
					t.Fatalf("Search calls = %d, want 1", repo.calls)
				}
				if err != repoErr {
					t.Errorf("Products() error = %v, want original error %v", err, repoErr)
				}
				if !errors.Is(err, tt.err) {
					t.Errorf("Products() error = %v, want errors.Is(_, %v)", err, tt.err)
				}
				if !reflect.DeepEqual(got, ProductsOutDTO{}) {
					t.Errorf("Products() = %#v, want zero DTO with partial data discarded", got)
				}
			})
		}
	}
}
