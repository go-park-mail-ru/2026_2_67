package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"

	"vibe_market/backend/internal/products/domain"
)

type ProductRepoPostgres struct {
	db *sql.DB
}

func NewProductRepo(db *sql.DB) *ProductRepoPostgres {
	return &ProductRepoPostgres{db: db}
}

func (r *ProductRepoPostgres) SelectByID(id int64, buyerID int64) (domain.ProductByID, error) {
	row := r.db.QueryRow(`
		SELECT p.id, p.category_id, p.seller_id, p.price, p.discount, p.name, p.description, p.available_count,
			p.created_at, p.updated_at, s.name, s.avatar_url, c.category,
			COALESCE((SELECT json_agg(pp.picture_url ORDER BY pp.id)
				FROM product_picture pp WHERE pp.product_id = p.id), '[]'::json),
			reviews.rating, reviews.reviews_count, reviews.marks_count,
			reviews.rating_1_count, reviews.rating_2_count, reviews.rating_3_count,
			reviews.rating_4_count, reviews.rating_5_count
		FROM product p
		JOIN category c ON c.id = p.category_id
		JOIN seller s ON s.user_id = p.seller_id
		CROSS JOIN LATERAL (
			SELECT COALESCE(AVG(pr.rating), 0)::double precision AS rating,
				COUNT(*) FILTER (WHERE pr.review IS NOT NULL) AS reviews_count,
				COUNT(*) AS marks_count,
				COUNT(*) FILTER (WHERE pr.rating = 1) AS rating_1_count,
				COUNT(*) FILTER (WHERE pr.rating = 2) AS rating_2_count,
				COUNT(*) FILTER (WHERE pr.rating = 3) AS rating_3_count,
				COUNT(*) FILTER (WHERE pr.rating = 4) AS rating_4_count,
				COUNT(*) FILTER (WHERE pr.rating = 5) AS rating_5_count
			FROM product_review pr WHERE pr.product_id = p.id
		) reviews
		WHERE p.id = $1`, id)

	productByID := domain.ProductByID{}
	var pictureURLsJSON []byte

	if row.Err() != nil {
		return productByID, row.Err()
	}

	err := row.Scan(&productByID.ID, &productByID.CategoryID, &productByID.SellerID,
		&productByID.Price, &productByID.Discount, &productByID.Name, &productByID.Description, &productByID.AvailableCount,
		&productByID.CreatedAt, &productByID.UpdatedAt, &productByID.SellerName, &productByID.SellerAvatarURL,
		&productByID.Category, &pictureURLsJSON, &productByID.Rating, &productByID.ReviewsCount,
		&productByID.MarksCount, &productByID.Rating1Count, &productByID.Rating2Count,
		&productByID.Rating3Count, &productByID.Rating4Count, &productByID.Rating5Count)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return productByID, domain.ErrProductIDNotFound
		}
		return productByID, err
	}
	if err := json.Unmarshal(pictureURLsJSON, &productByID.PictureURLS); err != nil {
		return productByID, err
	}

	// получил ли продукт покупатель
	row = r.db.QueryRow(`SELECT EXISTS (
		SELECT 1
		FROM order_product op
		JOIN "order" o ON op.order_id = o.id
		WHERE op.product_id = $1 AND o.order_status_id = $2 AND o.buyer_id = $3
	)`, id, domain.OrderStatusIDCanPostReview, buyerID)
	if row.Err() != nil {
		return domain.ProductByID{}, row.Err()
	}

	err = row.Scan(&productByID.CanReview)
	if err != nil {
		return domain.ProductByID{}, err
	}

	// отправлял ли он уже отзыв
	row = r.db.QueryRow(`SELECT EXISTS (
		SELECT 1 FROM product_review pr WHERE pr.product_id = $1 AND pr.buyer_id = $2
	)`, id, buyerID)
	if row.Err() != nil {
		return domain.ProductByID{}, row.Err()
	}

	var wasReview bool
	err = row.Scan(&wasReview)
	if err != nil {
		return domain.ProductByID{}, err
	}
	productByID.WasReview = wasReview
	if wasReview {
		productByID.CanReview = false
	}

	return productByID, nil
}
