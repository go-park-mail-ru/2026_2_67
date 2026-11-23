package postgres

import (
	"database/sql"
	"errors"

	"vibe_market/backend/internal/products/domain"
)

type ProductRepoPostgres struct {
	db *sql.DB
}

func NewProductRepo(db *sql.DB) *ProductRepoPostgres {
	return &ProductRepoPostgres{db: db}
}

func (r *ProductRepoPostgres) SelectByID(id int64) (domain.ProductByID, error) {
	row := r.db.QueryRow(`
		SELECT p.id, p.category_id, p.seller_id, p.price, p.discount, p.name, p.description, p.available_count,
    		p.created_at, p.updated_at, s.name, s.avatar_url, c.category, COALESCE(
        	(SELECT array_agg(pp.picture_url ORDER BY pp.id)
         		FROM product_picture pp
         		WHERE pp.product_id = p.id),
        		'{}')  AS picture_urls,
    COALESCE(AVG(pr.rating), 0)         AS rating,
    COUNT(pr.buyer_id)                  AS marks_count,
    COUNT(pr.review)                    AS reviews_count
	FROM product p
	JOIN category c ON c.id = p.category_id
	JOIN seller s ON s.user_id = p.seller_id
	LEFT JOIN product_review pr ON pr.product_id = p.id
	WHERE p.id = $1
	GROUP BY p.id, c.category, s.name, s.avatar_url;`, id)

	productByID := domain.ProductByID{}

	if row.Err() != nil {
		return productByID, row.Err()
	}

	err := row.Scan(&productByID.ID, &productByID.CategoryID, &productByID.SellerID,
		&productByID.Price, &productByID.Discount, &productByID.Name, &productByID.Description, &productByID.AvailableCount,
		&productByID.CreatedAt, &productByID.UpdatedAt, &productByID.SellerName, &productByID.SellerAvatarURL,
		&productByID.Category, &productByID.PictureURLS, &productByID.Rating,
		&productByID.MarksCount, &productByID.ReviewsCount)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return productByID, domain.ErrProductIDNotFound
		}
		return productByID, err
	}

	return productByID, nil
}
