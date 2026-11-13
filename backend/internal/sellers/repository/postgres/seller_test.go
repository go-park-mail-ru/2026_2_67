package postgres

import (
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"vibe_market/backend/internal/sellers/domain"
)

const sellerSelect = `SELECT s.user_id, s.name, s.description, s.avatar_url, s.created_at, s.updated_at
	FROM product p
	JOIN seller s ON p.seller_id = s.user_id
	JOIN category c ON p.category_id = c.id`

const sellerOrder = ` GROUP BY s.user_id ORDER BY COUNT(s.user_id) DESC, s.name ASC `

func sellerRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"user_id", "name", "description", "avatar_url", "created_at", "updated_at"})
}

func newSellerRepoTest(t *testing.T) (*SellerRepoPostgres, sqlmock.Sqlmock) {
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
	return NewSellerRepo(db), mock
}

func TestSellerSearchSQL(t *testing.T) {
	maxPrice, zero := 200.5, 0.0
	tests := []struct {
		name   string
		params domain.SellerSearchParams
		where  string
		page   string
		args   []driver.Value
	}{
		{
			name: "default limit", page: "OFFSET $1 LIMIT $2;",
			args: []driver.Value{0, 10},
		},
		{
			name: "explicit limit and offset", params: domain.SellerSearchParams{Limit: 7, Offset: 12},
			page: "OFFSET $1 LIMIT $2;", args: []driver.Value{12, 7},
		},
		{
			name: "limit at cap", params: domain.SellerSearchParams{Limit: 30},
			page: "OFFSET $1 LIMIT $2;", args: []driver.Value{0, 30},
		},
		{
			name: "limit above cap", params: domain.SellerSearchParams{Limit: 100},
			page: "OFFSET $1 LIMIT $2;", args: []driver.Value{0, 30},
		},
		{
			name: "category", params: domain.SellerSearchParams{Category: "Books"},
			where: " WHERE LOWER(c.category) = LOWER($1)", page: "OFFSET $2 LIMIT $3;",
			args: []driver.Value{"Books", 0, 10},
		},
		{
			name: "query escaping", params: domain.SellerSearchParams{Query: `50%_\'`},
			where: ` WHERE p.name ILIKE $1 ESCAPE E'\\'`, page: "OFFSET $2 LIMIT $3;",
			args: []driver.Value{`50\%\_\\'%`, 0, 10},
		},
		{
			name: "minimum only", params: domain.SellerSearchParams{PriceMin: 12.5},
			where: " WHERE p.price >= $1", page: "OFFSET $2 LIMIT $3;",
			args: []driver.Value{12.5, 0, 10},
		},
		{
			name: "zero maximum", params: domain.SellerSearchParams{PriceMax: &zero},
			where: " WHERE p.price <= $1", page: "OFFSET $2 LIMIT $3;",
			args: []driver.Value{0.0, 0, 10},
		},
		{
			name: "category and maximum", params: domain.SellerSearchParams{Category: "Books", PriceMax: &maxPrice},
			where: " WHERE LOWER(c.category) = LOWER($1) AND p.price <= $2", page: "OFFSET $3 LIMIT $4;",
			args: []driver.Value{"Books", 200.5, 0, 10},
		},
		{
			name: "query and minimum", params: domain.SellerSearchParams{Query: "Book", PriceMin: 10},
			where: ` WHERE p.name ILIKE $1 ESCAPE E'\\' AND p.price >= $2`, page: "OFFSET $3 LIMIT $4;",
			args: []driver.Value{"Book%", 10.0, 0, 10},
		},
		{
			name: "price range", params: domain.SellerSearchParams{PriceMin: 10, PriceMax: &maxPrice},
			where: " WHERE p.price >= $1 AND p.price <= $2", page: "OFFSET $3 LIMIT $4;",
			args: []driver.Value{10.0, 200.5, 0, 10},
		},
		{
			name: "all filters", params: domain.SellerSearchParams{
				Category: "Books", Query: `50%_\`, PriceMin: 10, PriceMax: &maxPrice, Offset: 9, Limit: 5,
			},
			where: ` WHERE LOWER(c.category) = LOWER($1) AND p.name ILIKE $2 ESCAPE E'\\' AND p.price >= $3 AND p.price <= $4`,
			page:  "OFFSET $5 LIMIT $6;", args: []driver.Value{"Books", `50\%\_\\%`, 10.0, 200.5, 9, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newSellerRepoTest(t)
			mock.ExpectQuery(sellerSelect + tt.where + sellerOrder + tt.page).
				WithArgs(tt.args...).WillReturnRows(sellerRows()).RowsWillBeClosed()

			got, err := repo.Search(tt.params)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, []domain.Seller{}) {
				t.Fatalf("Search() = %#v, want non-nil empty slice", got)
			}
		})
	}
}

func TestSellerSearchMapping(t *testing.T) {
	repo, mock := newSellerRepoTest(t)
	createdAt := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	description, avatar, empty := "description", "avatar.jpg", ""
	rows := sellerRows().
		AddRow(19, "Zulu", description, avatar, createdAt, updatedAt).
		AddRow(7, "Alpha", nil, nil, updatedAt, createdAt).
		AddRow(3, "Empty", empty, empty, createdAt, updatedAt)
	mock.ExpectQuery(sellerSelect+sellerOrder+"OFFSET $1 LIMIT $2;").
		WithArgs(0, 10).WillReturnRows(rows).RowsWillBeClosed()

	got, err := repo.Search(domain.SellerSearchParams{})
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Seller{
		{UserID: 19, Name: "Zulu", Description: &description, AvatarURL: &avatar, CreatedAt: createdAt, UpdatedAt: updatedAt},
		{UserID: 7, Name: "Alpha", CreatedAt: updatedAt, UpdatedAt: createdAt},
		{UserID: 3, Name: "Empty", Description: &empty, AvatarURL: &empty, CreatedAt: createdAt, UpdatedAt: updatedAt},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Search() = %#v, want %#v", got, want)
	}
}

func TestSellerSearchErrors(t *testing.T) {
	queryErr, rowsErr := errors.New("query failed"), errors.New("iteration failed")
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
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
			rows: sellerRows().AddRow(1, "Valid", nil, nil, now, now).
				AddRow("invalid", "Invalid", nil, nil, now, now).
				AddRow(3, "Unread", nil, nil, now, now),
			errText: "sql: Scan error on column index 0",
		},
		{
			name: "iteration error",
			rows: sellerRows().AddRow(1, "Valid", nil, nil, now, now).
				AddRow(2, "Invalid", nil, nil, now, now).RowError(1, rowsErr),
			wantErr: rowsErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newSellerRepoTest(t)
			expectation := mock.ExpectQuery(sellerSelect+sellerOrder+"OFFSET $1 LIMIT $2;").WithArgs(0, 10)
			if tt.queryErr != nil {
				expectation.WillReturnError(tt.queryErr)
			} else {
				expectation.WillReturnRows(tt.rows).RowsWillBeClosed()
			}

			got, err := repo.Search(domain.SellerSearchParams{})
			if err == nil {
				t.Fatal("Search() returned no error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.errText != "" && !strings.Contains(err.Error(), tt.errText) {
				t.Fatalf("error = %v, want containing %q", err, tt.errText)
			}
			if !reflect.DeepEqual(got, []domain.Seller{}) {
				t.Fatalf("Search() = %#v, want non-nil empty slice on error", got)
			}
		})
	}
}
