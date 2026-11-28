package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
	"vibe_market/backend/internal/products/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
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

func TestProductReviewSelectByProductIDAndBuyerID(t *testing.T) {
	now := time.Date(2026, 11, 24, 15, 0, 0, 0, time.UTC)
	reviewText := "Отлично"
	cols := []string{"buyer_id", "product_id", "review", "rating", "created_at", "updated_at"}

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT buyer_id, product_id, review, rating, created_at, updated_at`).
			WithArgs(int64(10), int64(34)).
			WillReturnRows(sqlmock.NewRows(cols).AddRow(int64(34), int64(10), &reviewText, 5, now, now))

		pr, err := NewProductReviewRepo(db).SelectByProductIDAndBuyerID(context.Background(), 10, 34)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pr.Rating != 5 || pr.Review == nil || *pr.Review != reviewText {
			t.Fatalf("unexpected review: %+v", pr)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		mock.ExpectQuery(`SELECT buyer_id, product_id, review, rating, created_at, updated_at`).
			WithArgs(int64(10), int64(34)).
			WillReturnError(sql.ErrNoRows)

		_, err = NewProductReviewRepo(db).SelectByProductIDAndBuyerID(context.Background(), 10, 34)
		if !errors.Is(err, domain.ErrProductReviewNotFound) {
			t.Fatalf("got error %v, want %v", err, domain.ErrProductReviewNotFound)
		}
	})
}

func TestProductReviewUpdate(t *testing.T) {
	now := time.Date(2026, 11, 24, 15, 0, 0, 0, time.UTC)
	newText := "Обновленный отзыв"
	cols := []string{"buyer_id", "product_id", "review", "rating", "created_at", "updated_at"}

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		mock.ExpectQuery(`UPDATE product_review`).
			WithArgs(int64(10), int64(34), &newText, 4).
			WillReturnRows(sqlmock.NewRows(cols).AddRow(int64(34), int64(10), &newText, 4, now, now))

		updated, err := NewProductReviewRepo(db).Update(context.Background(), domain.ProductReview{
			ProductID: 10, BuyerID: 34, Review: &newText, Rating: 4,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Rating != 4 || updated.Review == nil || *updated.Review != newText {
			t.Fatalf("unexpected updated review: %+v", updated)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		mock.ExpectQuery(`UPDATE product_review`).
			WithArgs(int64(10), int64(34), nil, 4).
			WillReturnError(sql.ErrNoRows)

		_, err = NewProductReviewRepo(db).Update(context.Background(), domain.ProductReview{
			ProductID: 10, BuyerID: 34, Review: nil, Rating: 4,
		})
		if !errors.Is(err, domain.ErrProductReviewNotFound) {
			t.Fatalf("got error %v, want %v", err, domain.ErrProductReviewNotFound)
		}
	})

	t.Run("invalid rating", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		mock.ExpectQuery(`UPDATE product_review`).
			WithArgs(int64(10), int64(34), nil, 9).
			WillReturnError(&pgconn.PgError{Code: "23514"})

		_, err = NewProductReviewRepo(db).Update(context.Background(), domain.ProductReview{
			ProductID: 10, BuyerID: 34, Review: nil, Rating: 9,
		})
		if !errors.Is(err, domain.ErrProductReviewRatingInvalid) {
			t.Fatalf("got error %v, want %v", err, domain.ErrProductReviewRatingInvalid)
		}
	})
}
