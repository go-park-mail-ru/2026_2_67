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
		"available_count", "created_at", "updated_at", "pictures", "rating", "review_count", "user_id", "seller_name", "avatar_url",
	})
}

func productRowsWithBuyCount() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "category_id", "seller_id", "price", "discount", "name", "description",
		"available_count", "buy_count", "created_at", "updated_at", "pictures", "rating", "review_count",
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

func expectPriceRange(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`(?s)SELECT COALESCE\(MIN\(p\.price.*FROM product p.*`).
		WillReturnRows(sqlmock.NewRows([]string{"min", "max"}).AddRow(0, 0))
}

func TestSearchReturnsProductsAndMetadata(t *testing.T) {
	repo, mock := newProductRepoTest(t)
	createdAt := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	description := "description"
	rows := productRows().
		AddRow(2, 3, 4, 100, 10, "Second", &description, 5, createdAt, createdAt, []byte(`[`+`"second.jpg"`+`]`), float64(4.5), int64(2), 4, "Seller", nil).
		AddRow(1, 3, 4, 50, 0, "First", nil, 0, createdAt.Add(-time.Hour), createdAt, []byte(`[]`), float64(0), int64(0), 4, "Seller", nil)
	expectPriceRange(mock)
	mock.ExpectQuery(`(?s)SELECT p\.id.*FROM product p.*ORDER BY p\.created_at DESC, p\.id DESC LIMIT \$3`).
		WithArgs(float64(0), 0, 2).
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
	if !reflect.DeepEqual(page.Ratings, []float64{4.5}) || !reflect.DeepEqual(page.ReviewCounts, []uint{2}) {
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
	expectPriceRange(mock)
	mock.ExpectQuery(`(?s)SELECT.*JOIN category c ON c\.id = p\.category_id.*JOIN seller s ON s\.user_id = p\.seller_id.*FROM product_review pr WHERE pr\.product_id = p\.id.*LOWER\(c\.category\) = LOWER\(\$1\).*LOWER\(s\.name\) IN \(\$2, \$3\).*p\.name ILIKE \$4.*p\.price.*\$5.*p\.price.*\$6.*\(p\.created_at, p\.id\) < \(\$8, \$9\).*LIMIT \$10`).
		WithArgs("Books", "seller", "other", `50\%\_\\%`, float64(10), float64(maxPrice), 0, createdAt, int64(7), 3).
		WillReturnRows(productRows())

	_, err := repo.Search(domain.ProductsSearchParams{
		Category: "Books", Seller: []string{"Seller", "Other"}, Query: query, PriceMin: 10,
		PriceMax: &maxPrice, Cursor: &domain.Cursor{CreatedAt: createdAt, ID: 7}, Limit: 2,
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
		{"price", domain.ProductsSearchParams{PriceMin: 11, PriceMax: &maxPrice}, domain.ErrPriceRangeInvalid},
		{"cursor", domain.ProductsSearchParams{Cursor: &domain.Cursor{ID: 1}}, domain.ErrCursorInvalid},
		{"sort", domain.ProductsSearchParams{SortBy: "other"}, domain.ErrSortByInvalid},
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
		AddRow(2, 3, 4, 100.0, 10, "A", nil, 1, createdAt, createdAt, []byte(`[]`), rating, int64(3), 4, "Seller", nil).
		AddRow(1, 3, 4, 100.0, 10, "B", nil, 1, createdAt, createdAt, []byte(`[]`), rating, int64(3), 4, "Seller", nil)
	expectPriceRange(mock)
	mock.ExpectQuery(`(?s)ORDER BY reviews\.rating DESC, p\.id DESC LIMIT \$3`).
		WithArgs(float64(0), 0, 2).WillReturnRows(rows)
	page, err := repo.Search(domain.ProductsSearchParams{SortBy: domain.SortByHighRating, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	want := &domain.Cursor{ID: 2, Rating: rating}
	if !reflect.DeepEqual(page.NextCursor, want) {
		t.Fatalf("got cursor %+v, want %+v", page.NextCursor, want)
	}

	expectPriceRange(mock)
	mock.ExpectQuery(`(?s)\(reviews\.rating, p\.id\) < \(\$3, \$4\).*ORDER BY reviews\.rating DESC, p\.id DESC LIMIT \$5`).
		WithArgs(float64(0), 0, rating, int64(2), 2).WillReturnRows(productRows())
	_, err = repo.Search(domain.ProductsSearchParams{SortBy: domain.SortByHighRating, Limit: 1, Cursor: page.NextCursor})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchReturnsDatabaseError(t *testing.T) {
	repo, mock := newProductRepoTest(t)
	dbErr := errors.New("database unavailable")
	expectPriceRange(mock)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT p.id")).
		WithArgs(float64(0), 0, productsDefaultLimit+1).
		WillReturnError(dbErr)

	_, err := repo.Search(domain.ProductsSearchParams{})
	if !errors.Is(err, dbErr) {
		t.Fatalf("got %v, want %v", err, dbErr)
	}
}

func TestSearchReturnsRowsError(t *testing.T) {
	repo, mock := newProductRepoTest(t)
	rowsErr := errors.New("rows failed")
	rows := productRows().AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).RowError(0, rowsErr)
	expectPriceRange(mock)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT p.id")).
		WithArgs(float64(0), 0, productsDefaultLimit+1).
		WillReturnRows(rows)

	_, err := repo.Search(domain.ProductsSearchParams{})
	if !errors.Is(err, rowsErr) {
		t.Fatalf("got %v, want %v", err, rowsErr)
	}
}
