package postgres

import (
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"vibe_market/backend/internal/categories/domain"
)

const categoryNewQuery = "SELECT id, category, picture_url FROM category ORDER BY id DESC"

const categoryPopularQuery = `SELECT c.id, c.category, c.picture_url FROM category c LEFT JOIN product p ON p.category_id = c.id
	GROUP BY c.id, c.category, c.picture_url ORDER BY COUNT(p.id) DESC, c.id DESC`

func categoryRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "category", "picture_url"})
}

func newCategoryRepoTest(t *testing.T) (*CategoryRepoPostgres, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		_ = db.Close()
	})
	return NewCategoryRepo(db), mock
}

func TestCategorySearchSQL(t *testing.T) {
	tests := []struct {
		name   string
		params domain.CategoriesSearchParams
		query  string
		args   []driver.Value
	}{
		{
			name: "default without limit", query: categoryNewQuery + ";",
		},
		{
			name: "default with limit", params: domain.CategoriesSearchParams{Limit: 7},
			query: categoryNewQuery + " LIMIT $1;", args: []driver.Value{7},
		},
		{
			name: "new without limit", params: domain.CategoriesSearchParams{SortBy: domain.SortByNew},
			query: categoryNewQuery + ";",
		},
		{
			name: "new with limit", params: domain.CategoriesSearchParams{SortBy: domain.SortByNew, Limit: 3},
			query: categoryNewQuery + " LIMIT $1;", args: []driver.Value{3},
		},
		{
			name: "popular without limit", params: domain.CategoriesSearchParams{SortBy: domain.SortByPopular},
			query: categoryPopularQuery + " ;",
		},
		{
			name: "popular with limit", params: domain.CategoriesSearchParams{SortBy: domain.SortByPopular, Limit: 5},
			query: categoryPopularQuery + " LIMIT $1;", args: []driver.Value{5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newCategoryRepoTest(t)
			mock.ExpectQuery(tt.query).WithArgs(tt.args...).
				WillReturnRows(categoryRows()).RowsWillBeClosed()

			got, err := repo.Search(tt.params)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, []domain.Category{}) {
				t.Fatalf("Search() = %#v, want non-nil empty slice", got)
			}
		})
	}
}

func TestCategorySearchInvalidSort(t *testing.T) {
	repo, _ := newCategoryRepoTest(t)
	// No query is expected: invalid sorting must be rejected before accessing the DB.
	got, err := repo.Search(domain.CategoriesSearchParams{SortBy: "invalid", Limit: 5})
	if !errors.Is(err, domain.ErrInvalidSortBy) {
		t.Fatalf("error = %v, want %v", err, domain.ErrInvalidSortBy)
	}
	if !reflect.DeepEqual(got, []domain.Category{}) {
		t.Fatalf("Search() = %#v, want non-nil empty slice", got)
	}
}

func TestCategorySearchMapping(t *testing.T) {
	repo, mock := newCategoryRepoTest(t)
	picture, empty := "category.jpg", ""
	rows := categoryRows().
		AddRow(19, "Books", picture).
		AddRow(7, "Clothes", nil).
		AddRow(3, "Food", empty)
	mock.ExpectQuery(categoryNewQuery + ";").WithArgs().
		WillReturnRows(rows).RowsWillBeClosed()

	got, err := repo.Search(domain.CategoriesSearchParams{})
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Category{
		{ID: 19, Category: "Books", PictureURL: &picture},
		{ID: 7, Category: "Clothes"},
		{ID: 3, Category: "Food", PictureURL: &empty},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Search() = %#v, want %#v", got, want)
	}
}

func TestCategorySearchErrors(t *testing.T) {
	queryErr, rowsErr := errors.New("query failed"), errors.New("iteration failed")
	tests := []struct {
		name     string
		rows     *sqlmock.Rows
		queryErr error
		wantErr  error
		errText  string
	}{
		{name: "query error", queryErr: queryErr, wantErr: queryErr},
		{
			name: "scan error",
			rows: categoryRows().AddRow(1, "Valid", nil).
				AddRow("invalid", "Invalid", nil).
				AddRow(3, "Unread", nil),
			errText: "sql: Scan error on column index 0",
		},
		{
			name: "iteration error",
			rows: categoryRows().AddRow(1, "Valid", nil).
				AddRow(2, "Invalid", nil).RowError(1, rowsErr),
			wantErr: rowsErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newCategoryRepoTest(t)
			expectation := mock.ExpectQuery(categoryNewQuery + ";").WithArgs()
			if tt.queryErr != nil {
				expectation.WillReturnError(tt.queryErr)
			} else {
				expectation.WillReturnRows(tt.rows).RowsWillBeClosed()
			}

			got, err := repo.Search(domain.CategoriesSearchParams{})
			if err == nil {
				t.Fatal("Search() returned no error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.errText != "" && !strings.Contains(err.Error(), tt.errText) {
				t.Fatalf("error = %v, want containing %q", err, tt.errText)
			}
			if !reflect.DeepEqual(got, []domain.Category{}) {
				t.Fatalf("Search() = %#v, want non-nil empty slice on error", got)
			}
		})
	}
}
