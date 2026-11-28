package postgres

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"vibe_market/backend/internal/products/domain"
)

func TestProductReviewInsert(t *testing.T) {
	tests := []struct {
		name  string
		dbErr error
		want  error
	}{
		{name: "success"},
		{name: "duplicate", dbErr: &pgconn.PgError{Code: "23505"}, want: domain.ErrProductReviewAlreadyExists},
		{name: "invalid rating", dbErr: &pgconn.PgError{Code: "23514"}, want: domain.ErrProductReviewRatingInvalid},
		{name: "database error", dbErr: sqlmock.ErrCancelled, want: sqlmock.ErrCancelled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			query := mock.ExpectExec(`INSERT INTO product_review`).WithArgs(int64(34), int64(12), "Good", 5)
			if tt.dbErr != nil {
				query.WillReturnError(tt.dbErr)
			} else {
				query.WillReturnResult(sqlmock.NewResult(0, 1))
			}
			text := "Good"
			err = NewProductReviewRepo(db).Insert(domain.ProductReview{BuyerID: 34, ProductID: 12, Review: &text, Rating: 5})
			if !errors.Is(err, tt.want) {
				t.Fatalf("got error %v, want %v", err, tt.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
