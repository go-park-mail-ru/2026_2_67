package postgres

import (
	"context"
	"errors"
	"testing"
	"time"
	"vibe_market/backend/internal/products/domain"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestProductReviewSelectByProductID(t *testing.T) {
	now := time.Date(2026, 11, 24, 15, 0, 0, 0, time.UTC)
	reviewText := "Отличный товар"
	surname := "Иванов"
	patronymic := "Иванович"
	avatarURL := "https://cdn.example.com/a.jpg"

	columns := []string{"buyer_id", "name", "surname", "patronymic", "avatar_url", "review", "rating", "created_at"}

	t.Run("success with next cursor", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		rows := sqlmock.NewRows(columns).
			AddRow(int64(1), "Иван", &surname, &patronymic, &avatarURL, &reviewText, 5, now).
			AddRow(int64(2), "Пётр", nil, nil, nil, nil, 4, now.Add(-time.Hour)).
			AddRow(int64(3), "Анна", nil, nil, nil, nil, 3, now.Add(-2*time.Hour))

		mock.ExpectQuery(`SELECT pr\.buyer_id, b\.name, b\.surname, b\.patronymic, b\.avatar_url, pr\.review, pr\.rating, pr\.created_at`).
			WithArgs(int64(10), 3).
			WillReturnRows(rows)

		page, err := NewProductReviewRepo(db).SelectByProductID(context.Background(), domain.ProductReviewsParams{
			ProductID: 10,
			SortBy:    domain.ReviewSortByDescDate,
			Limit:     2,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(page.Reviews) != 2 {
			t.Fatalf("got %d reviews, want 2", len(page.Reviews))
		}
		if page.NextCursor == nil || page.NextCursor.BuyerID != 2 || page.NextCursor.Rating != 4 {
			t.Fatalf("unexpected next cursor: %+v", page.NextCursor)
		}
		if page.Reviews[0].BuyerName != "Иван" || *page.Reviews[0].Review != reviewText {
			t.Fatalf("unexpected first review: %+v", page.Reviews[0])
		}
		if page.Reviews[1].Review != nil || page.Reviews[1].BuyerSurname != nil {
			t.Fatalf("expected nil nullable fields, got %+v", page.Reviews[1])
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("asc-rating with cursor", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		cursor := &domain.ProductReviewCursor{CreatedAt: now, BuyerID: 5, Rating: 2}
		rows := sqlmock.NewRows(columns).
			AddRow(int64(7), "Олег", nil, nil, nil, nil, 3, now)

		mock.ExpectQuery(`ORDER BY pr\.rating ASC, pr\.created_at DESC, pr\.buyer_id DESC`).
			WithArgs(int64(10), 11, 2, now, int64(5)).
			WillReturnRows(rows)

		page, err := NewProductReviewRepo(db).SelectByProductID(context.Background(), domain.ProductReviewsParams{
			ProductID: 10,
			SortBy:    domain.ReviewSortByAscRating,
			Limit:     10,
			Cursor:    cursor,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(page.Reviews) != 1 || page.NextCursor != nil {
			t.Fatalf("unexpected page: %+v", page)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT pr\.buyer_id`).
			WithArgs(int64(10), 11).
			WillReturnError(sqlmock.ErrCancelled)

		_, err = NewProductReviewRepo(db).SelectByProductID(context.Background(), domain.ProductReviewsParams{
			ProductID: 10,
			SortBy:    domain.ReviewSortByDescDate,
			Limit:     10,
		})
		if !errors.Is(err, sqlmock.ErrCancelled) {
			t.Fatalf("got error %v, want %v", err, sqlmock.ErrCancelled)
		}
	})
}
