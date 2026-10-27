package postgres

import (
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"vibe_market/backend/internal/products/domain"
)

func productRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "category_id", "seller_id", "price", "discount", "name", "description",
		"available_count", "created_at", "updated_at", "pictures", "rating", "review_count",
	})
}

func newProductRepoTest(t *testing.T) (*ProductRepoPostgres, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		_ = db.Close()
	})
	return NewProductRepo(db), mock
}

func TestSearchReturnsProductsAndMetadata(t *testing.T) {
	repo, mock := newProductRepoTest(t)
	createdAt := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	description := "description"
	rows := productRows().
		AddRow(2, 3, 4, 100.5, 10, "Second", &description, 5, createdAt, createdAt, []byte(`[`+`"second.jpg"`+`]`), float64(4.5), int64(2)).
		AddRow(1, 3, 4, 50.25, 0, "First", nil, 0, createdAt.Add(-time.Hour), createdAt, []byte(`[]`), float64(0), int64(0))
	mock.ExpectQuery(`(?s)SELECT p\.id.*FROM product p.*ORDER BY p\.created_at DESC, p\.id DESC LIMIT \$2`).
		WithArgs(float64(0), 2).
		WillReturnRows(rows)

	page, err := repo.Search(domain.ProductsSearchParams{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Products) != 1 || page.Products[0].ID != 2 {
		t.Fatalf("unexpected products: %+v", page.Products)
	}
	if !reflect.DeepEqual(page.PictureURLs, [][]string{{"second.jpg"}}) {
		t.Fatalf("unexpected pictures: %#v", page.PictureURLs)
	}
	if !reflect.DeepEqual(page.Ratings, []float32{4.5}) || !reflect.DeepEqual(page.ReviewCounts, []uint{2}) {
		t.Fatalf("unexpected review metadata: ratings=%v counts=%v", page.Ratings, page.ReviewCounts)
	}
	wantCursor := &domain.Cursor{CreatedAt: createdAt, ID: 2}
	if !reflect.DeepEqual(page.NextCursor, wantCursor) {
		t.Fatalf("unexpected next cursor: %+v", page.NextCursor)
	}
}

func TestSearchBuildsFiltersAndCursor(t *testing.T) {
	repo, mock := newProductRepoTest(t)
	maxPrice := 200.5
	createdAt := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	query := `50%_\`
	mock.ExpectQuery(`(?s)SELECT.*LOWER\(c\.category\) = LOWER\(\$1\).*LOWER\(s\.name\) = LOWER\(\$2\).*p\.name ILIKE \$3.*p\.price >= \$4.*p\.price <= \$5.*\(p\.created_at, p\.id\) < \(\$6, \$7\).*LIMIT \$8`).
		WithArgs("Books", "Seller", `50\%\_\\%`, float64(10), float64(maxPrice), createdAt, int64(7), 3).
		WillReturnRows(productRows())

	_, err := repo.Search(domain.ProductsSearchParams{
		Category: "Books", Seller: "Seller", Query: query, PriceMin: 10,
		PriceMax: &maxPrice, Cursor: domain.Cursor{CreatedAt: createdAt, ID: 7}, Limit: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchRejectsInvalidParameters(t *testing.T) {
	maxPrice := 10.0
	tests := []struct {
		name   string
		params domain.ProductsSearchParams
		want   error
	}{
		{"price", domain.ProductsSearchParams{PriceMin: 11, PriceMax: &maxPrice}, domain.ErrInvalidPriceRange},
		{"cursor", domain.ProductsSearchParams{Cursor: domain.Cursor{ID: 1}}, domain.ErrInvalidCursor},
		{"sort", domain.ProductsSearchParams{SortBy: "other"}, domain.ErrInvalidSortBy},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewProductRepo(nil).Search(test.params)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestSearchPopularPagination(t *testing.T) {
	repo, mock := newProductRepoTest(t)
	createdAt := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	rating := float64(4.123456789)
	rows := productRows().
		AddRow(2, 3, 4, 100.0, 10, "A", nil, 1, createdAt, createdAt, []byte(`[]`), rating, int64(3)).
		AddRow(1, 3, 4, 100.0, 10, "B", nil, 1, createdAt, createdAt, []byte(`[]`), rating, int64(3))
	mock.ExpectQuery(`(?s)ORDER BY reviews\.rating DESC, p\.created_at DESC, p\.id DESC LIMIT \$2`).
		WithArgs(float64(0), 2).WillReturnRows(rows)
	page, err := repo.Search(domain.ProductsSearchParams{SortBy: domain.SortByPopular, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	want := &domain.Cursor{CreatedAt: createdAt, ID: 2, Rating: rating}
	if !reflect.DeepEqual(page.NextCursor, want) {
		t.Fatalf("got cursor %+v, want %+v", page.NextCursor, want)
	}

	mock.ExpectQuery(`(?s)\(reviews\.rating, p\.created_at, p\.id\) < \(\$2, \$3, \$4\).*ORDER BY reviews\.rating DESC, p\.created_at DESC, p\.id DESC LIMIT \$5`).
		WithArgs(float64(0), rating, createdAt, int64(2), 2).WillReturnRows(productRows())
	_, err = repo.Search(domain.ProductsSearchParams{SortBy: domain.SortByPopular, Limit: 1, Cursor: *page.NextCursor})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchReturnsDatabaseError(t *testing.T) {
	repo, mock := newProductRepoTest(t)
	dbErr := errors.New("database unavailable")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT p.id")).
		WithArgs(float64(0), ProductsDefaultLimit+1).
		WillReturnError(dbErr)

	_, err := repo.Search(domain.ProductsSearchParams{})
	if !errors.Is(err, dbErr) {
		t.Fatalf("got %v, want %v", err, dbErr)
	}
}

func TestSearchReturnsRowsError(t *testing.T) {
	repo, mock := newProductRepoTest(t)
	rowsErr := errors.New("rows failed")
	rows := productRows().AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).RowError(0, rowsErr)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT p.id")).
		WithArgs(float64(0), ProductsDefaultLimit+1).
		WillReturnRows(rows)

	_, err := repo.Search(domain.ProductsSearchParams{})
	if !errors.Is(err, rowsErr) {
		t.Fatalf("got %v, want %v", err, rowsErr)
	}
}
