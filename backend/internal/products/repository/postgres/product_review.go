package postgres

import (
	"database/sql"
	"errors"
	"vibe_market/backend/internal/products/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

type ProductReviewRepoPostgres struct {
	db *sql.DB
}

func NewProductReviewRepo(db *sql.DB) *ProductReviewRepoPostgres {
	return &ProductReviewRepoPostgres{db: db}
}

func (repo *ProductReviewRepoPostgres) Insert(pr domain.ProductReview) error {
	_, err := repo.db.Exec(`INSERT INTO product_review (buyer_id, product_id, review, rating) VALUES ($1, $2, $3, $4)`,
		pr.BuyerID, pr.ProductID, pr.Review, pr.Rating)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrProductReviewAlreadyExists
		}
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			return domain.ErrProductReviewRatingInvalid
		}
		return err
	}

	return nil
}
