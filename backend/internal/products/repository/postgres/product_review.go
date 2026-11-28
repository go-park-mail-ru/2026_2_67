package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"vibe_market/backend/internal/products/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

type ProductReviewRepoPostgres struct {
	db *sql.DB
}

func NewProductReviewRepo(db *sql.DB) *ProductReviewRepoPostgres {
	return &ProductReviewRepoPostgres{db: db}
}

func (r *ProductReviewRepoPostgres) Insert(pr domain.ProductReview) error {
	_, err := r.db.Exec(`INSERT INTO product_review (buyer_id, product_id, review, rating) VALUES ($1, $2, $3, $4)`,
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

func reviewSortOrderBy(sortBy domain.ProductReviewSortBy) string {
	switch sortBy {
	case domain.ReviewSortByAscDate:
		return "pr.created_at ASC, pr.buyer_id ASC"
	case domain.ReviewSortByDescRating:
		return "pr.rating DESC, pr.created_at DESC, pr.buyer_id DESC"
	case domain.ReviewSortByAscRating:
		return "pr.rating ASC, pr.created_at DESC, pr.buyer_id DESC"
	default:
		return "pr.created_at DESC, pr.buyer_id DESC"
	}
}

func reviewCursorCondition(sortBy domain.ProductReviewSortBy, cursor *domain.ProductReviewCursor, args []any) (string, []any) {
	if cursor == nil {
		return "", args
	}

	switch sortBy {
	case domain.ReviewSortByAscDate:
		args = append(args, cursor.CreatedAt, cursor.BuyerID)
		return fmt.Sprintf("AND (pr.created_at, pr.buyer_id) > ($%d, $%d)", len(args)-1, len(args)), args
	case domain.ReviewSortByDescRating:
		args = append(args, cursor.Rating, cursor.CreatedAt, cursor.BuyerID)
		return fmt.Sprintf("AND (pr.rating, pr.created_at, pr.buyer_id) < ($%d, $%d, $%d)", len(args)-2, len(args)-1, len(args)), args
	case domain.ReviewSortByAscRating:
		args = append(args, cursor.Rating, cursor.CreatedAt, cursor.BuyerID)
		return fmt.Sprintf("AND (pr.rating > $%d OR (pr.rating = $%d AND (pr.created_at, pr.buyer_id) < ($%d, $%d)))",
			len(args)-2, len(args)-2, len(args)-1, len(args)), args
	default:
		args = append(args, cursor.CreatedAt, cursor.BuyerID)
		return fmt.Sprintf("AND (pr.created_at, pr.buyer_id) < ($%d, $%d)", len(args)-1, len(args)), args
	}
}

func (r *ProductReviewRepoPostgres) SelectByProductID(ctx context.Context, params domain.ProductReviewsParams) (domain.ProductReviewsPage, error) {
	args := []any{params.ProductID, params.Limit + 1}
	cursorClause, args := reviewCursorCondition(params.SortBy, params.Cursor, args)
	orderBy := reviewSortOrderBy(params.SortBy)

	query := fmt.Sprintf(`
		SELECT pr.buyer_id, b.name, b.surname, b.patronymic, b.avatar_url, pr.review, pr.rating, pr.created_at
		FROM product_review pr
		JOIN buyer b ON b.user_id = pr.buyer_id
		WHERE pr.product_id = $1
		%s
		ORDER BY %s
		LIMIT $2
	`, cursorClause, orderBy)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.ProductReviewsPage{}, err
	}
	defer rows.Close()

	reviews := make([]domain.ProductReviewListItem, 0, params.Limit)
	for rows.Next() {
		var item domain.ProductReviewListItem
		if err := rows.Scan(
			&item.BuyerID,
			&item.BuyerName,
			&item.BuyerSurname,
			&item.BuyerPatronymic,
			&item.BuyerAvatarURL,
			&item.Review,
			&item.Rating,
			&item.CreatedAt,
		); err != nil {
			return domain.ProductReviewsPage{}, err
		}
		reviews = append(reviews, item)
	}
	if err := rows.Err(); err != nil {
		return domain.ProductReviewsPage{}, err
	}

	page := domain.ProductReviewsPage{Reviews: reviews}
	if len(reviews) > params.Limit {
		page.Reviews = reviews[:params.Limit]
		last := page.Reviews[len(page.Reviews)-1]
		page.NextCursor = &domain.ProductReviewCursor{
			CreatedAt: last.CreatedAt,
			BuyerID:   last.BuyerID,
			Rating:    last.Rating,
		}
	}

	return page, nil
}

func (r *ProductReviewRepoPostgres) SelectByProductIDAndBuyerID(ctx context.Context, productID, buyerID int64) (domain.ProductReview, error) {
	var pr domain.ProductReview
	err := r.db.QueryRowContext(
		ctx,
		`SELECT buyer_id, product_id, review, rating, created_at, updated_at
		 FROM product_review
		 WHERE product_id = $1 AND buyer_id = $2`,
		productID, buyerID,
	).Scan(&pr.BuyerID, &pr.ProductID, &pr.Review, &pr.Rating, &pr.CreatedAt, &pr.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProductReview{}, domain.ErrProductReviewNotFound
		}
		return domain.ProductReview{}, err
	}
	return pr, nil
}

func (r *ProductReviewRepoPostgres) Update(ctx context.Context, pr domain.ProductReview) (domain.ProductReview, error) {
	var updated domain.ProductReview
	err := r.db.QueryRowContext(
		ctx,
		`UPDATE product_review
		 SET review = $3, rating = $4
		 WHERE product_id = $1 AND buyer_id = $2
		 RETURNING buyer_id, product_id, review, rating, created_at, updated_at`,
		pr.ProductID, pr.BuyerID, pr.Review, pr.Rating,
	).Scan(&updated.BuyerID, &updated.ProductID, &updated.Review, &updated.Rating, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProductReview{}, domain.ErrProductReviewNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			return domain.ProductReview{}, domain.ErrProductReviewRatingInvalid
		}
		return domain.ProductReview{}, err
	}
	return updated, nil
}
