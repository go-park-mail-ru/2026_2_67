package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"vibe_market/backend/internal/products/domain"
)

type ProductRepoPostgres struct {
	db *sql.DB
}

var _ domain.ProductRepo = (*ProductRepoPostgres)(nil)

func NewProductRepo(db *sql.DB) *ProductRepoPostgres {
	return &ProductRepoPostgres{db: db}
}

// Search возвращает товары от новых к старым, используя id для одинаковых created_at.
// Нулевой Cursor означает первую страницу.
func (repo *ProductRepoPostgres) Search(params domain.SearchParams) (domain.ProductsPage, error) {
	if params.Limit > domain.MaxLimit {
		return domain.ProductsPage{}, domain.ErrLimitOverflow
	}
	invalidPrice := func(price float32) bool {
		return price < 0 || math.IsNaN(float64(price)) || math.IsInf(float64(price), 0)
	}
	if invalidPrice(params.PriceMin) || params.PriceMax != nil && (invalidPrice(*params.PriceMax) || params.PriceMin > *params.PriceMax) {
		return domain.ProductsPage{}, domain.ErrInvalidPriceRange
	}
	hasCursor := params.Cursor != (domain.Cursor{})
	if hasCursor && (params.Cursor.CreatedAt.IsZero() || params.Cursor.ID <= 0) {
		return domain.ProductsPage{}, domain.ErrInvalidCursor
	}

	limit := params.Limit
	if limit == 0 {
		limit = domain.DefaultLimit
	}

	conditions := make([]string, 0, 6)
	args := make([]any, 0, 8)
	addCondition := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if params.Category != "" {
		addCondition("LOWER(c.category) = LOWER($%d)", params.Category)
	}
	if params.Seller != "" {
		addCondition("LOWER(s.name) = LOWER($%d)", params.Seller)
	}
	if params.Query != "" {
		// LIKE-метасимволы пользовательского ввода должны искаться буквально.
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(params.Query)
		addCondition(`p.name ILIKE $%d ESCAPE E'\\'`, escaped+"%")
	}
	// Цена в БД целочисленная, но границы фильтра могут быть дробными.
	addCondition("p.price >= $%d::double precision", float64(params.PriceMin))
	if params.PriceMax != nil {
		addCondition("p.price <= $%d::double precision", float64(*params.PriceMax))
	}
	if hasCursor {
		args = append(args, params.Cursor.CreatedAt, params.Cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(p.created_at, p.id) < ($%d, $%d)", len(args)-1, len(args)))
	}

	query := `SELECT p.id, p.category_id, p.seller_id, p.price, p.name,
		p.description, p.available_count, p.created_at, p.updated_at,
		COALESCE((SELECT json_agg(pp.picture_url ORDER BY pp.created_at, pp.id)
			FROM product_picture pp WHERE pp.product_id = p.id), '[]'::json),
		COALESCE(reviews.rating, 0), reviews.review_count
		FROM product p
		JOIN product_category c ON c.id = p.category_id
		JOIN seller s ON s.user_id = p.seller_id
		LEFT JOIN LATERAL (
			SELECT AVG(pr.rating) AS rating, COUNT(*) AS review_count
			FROM product_review pr WHERE pr.product_id = p.id
		) reviews ON TRUE
		WHERE ` + strings.Join(conditions, " AND ")
	args = append(args, int(limit)+1)
	query += fmt.Sprintf(" ORDER BY p.created_at DESC, p.id DESC LIMIT $%d", len(args))

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return domain.ProductsPage{}, fmt.Errorf("search products: %w", err)
	}
	defer rows.Close()

	page := domain.ProductsPage{
		Products:     make([]domain.Product, 0, int(limit)),
		PictureURLs:  make([][]string, 0, int(limit)),
		Ratings:      make([]float32, 0, int(limit)),
		ReviewCounts: make([]uint, 0, int(limit)),
	}
	for rows.Next() {
		var (
			product      domain.Product
			picturesJSON []byte
			rating       float32
			reviewCount  uint
		)
		if err := rows.Scan(
			&product.ID, &product.CategoryID, &product.SellerID, &product.Price,
			&product.Name, &product.Description, &product.AvailableCount,
			&product.CreatedAt, &product.UpdatedAt,
			&picturesJSON, &rating, &reviewCount,
		); err != nil {
			return domain.ProductsPage{}, fmt.Errorf("scan product: %w", err)
		}
		if len(page.Products) < int(limit) {
			var pictureURLs []string
			if err := json.Unmarshal(picturesJSON, &pictureURLs); err != nil {
				return domain.ProductsPage{}, fmt.Errorf("decode product pictures: %w", err)
			}
			page.Products = append(page.Products, product)
			page.PictureURLs = append(page.PictureURLs, pictureURLs)
			page.Ratings = append(page.Ratings, rating)
			page.ReviewCounts = append(page.ReviewCounts, reviewCount)
		} else {
			last := page.Products[len(page.Products)-1]
			page.NextCursor = &domain.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
		}
	}
	if err := rows.Err(); err != nil {
		return domain.ProductsPage{}, fmt.Errorf("read products: %w", err)
	}
	return page, nil
}
