package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"vibe_market/backend/internal/products/domain"
	sellerDomain "vibe_market/backend/internal/sellers/domain"
)

type ProductRepoPostgres struct {
	db *sql.DB
}

func NewProductRepo(db *sql.DB) *ProductRepoPostgres {
	return &ProductRepoPostgres{db: db}
}

func isPriceInvalid(price float64) bool {
	return price < 0 || math.IsNaN(price) || math.IsInf(price, 0)
}

const productsDefaultLimit = 20
const productsMaxLimit = 2 * productsDefaultLimit

func AddLastViewedProductByCursor(cursor domain.Cursor, sortBy domain.SortByType, args []any, conditions []string) ([]any, []string) {
	switch sortBy {
	case domain.SortByNew:
		args = append(args, cursor.CreatedAt, cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(p.created_at, p.id) < ($%d, $%d)", len(args)-1, len(args)))
	case domain.SortByAscPrice:
		args = append(args, cursor.Price, cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(p.price > $%d OR (p.price = $%d AND p.id < $%d))", len(args)-1, len(args)-1, len(args)))
	case domain.SortByDescPrice:
		args = append(args, cursor.Price, cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(p.price, p.id) < ($%d, $%d)", len(args)-1, len(args)))
	case domain.SortByBuyCount:
		args = append(args, cursor.BuyCount, cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(p.buy_count, p.id) < ($%d, $%d)", len(args)-1, len(args)))
	case domain.SortByHighRating:
		args = append(args, cursor.Rating, cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(reviews.rating, p.id) < ($%d, $%d)", len(args)-1, len(args)))
	}
	return args, conditions
}

func SortingBySortBy(sortBy domain.SortByType) string {
	switch sortBy {
	case domain.SortByNew:
		return "p.created_at DESC, p.id DESC"
	case domain.SortByAscPrice:
		return "p.price ASC, p.id DESC"
	case domain.SortByDescPrice:
		return "p.price DESC, p.id DESC"
	case domain.SortByBuyCount:
		return "p.buy_count DESC, p.id DESC"
	case domain.SortByHighRating:
		return "reviews.rating DESC, p.id DESC"
	default:
		return ""
	}
}

// WARN: медленный. Зачем на buyCount при sortBy == new???

func (r *ProductRepoPostgres) Search(params domain.ProductsSearchParams) (domain.ProductsPage, error) {
	if params.SortBy == "" {
		params.SortBy = domain.SortByNew
	}
	if params.Limit > productsMaxLimit {
		params.Limit = productsMaxLimit
	}
	if isPriceInvalid(params.PriceMin) || params.PriceMax != nil &&
		(isPriceInvalid(*params.PriceMax) || params.PriceMin > *params.PriceMax) {
		return domain.ProductsPage{}, domain.ErrPriceRangeInvalid
	}
	hasCursor := params.Cursor != nil
	if hasCursor && (params.Cursor.CreatedAt.IsZero() || params.Cursor.ID <= 0) {
		return domain.ProductsPage{}, domain.ErrCursorInvalid
	}

	if params.Limit == 0 {
		params.Limit = productsDefaultLimit
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
	if len(params.Seller) != 0 {
		placeholders := make([]string, 0, len(params.Seller))
		for _, seller := range params.Seller {
			args = append(args, strings.ToLower(seller))
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
		}
		conditions = append(conditions, "LOWER(s.name) IN ("+strings.Join(placeholders, ", ")+")")
	}
	if params.Query != "" {
		// LIKE-метасимволы пользовательского ввода должны искаться буквально.
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(params.Query)
		addCondition(`p.name ILIKE $%d ESCAPE E'\\'`, escaped+"%")
	}
	addCondition("p.price >= $%d", params.PriceMin)
	if params.PriceMax != nil {
		addCondition("p.price <= $%d", *params.PriceMax)
	}
	if hasCursor {
		args, conditions = AddLastViewedProductByCursor(*params.Cursor, params.SortBy, args, conditions)
	}

	addCondition("p.available_count > $%d", 0)
	query := `SELECT p.id, p.category_id, p.seller_id, p.price, p.discount, p.name,
		p.description, p.available_count, p.buy_count, p.created_at, p.updated_at,
		COALESCE((SELECT json_agg(pp.picture_url ORDER BY pp.created_at, pp.id)
			FROM product_picture pp WHERE pp.product_id = p.id), '[]'::json),
		reviews.rating, reviews.review_count, s.user_id, s.name, s.avatar_url
		FROM (
			SELECT p.*, COALESCE(SUM(op.count) FILTER (WHERE o.order_status_id = 4), 0) AS buy_count
			FROM product p
			LEFT JOIN order_product op ON op.product_id = p.id
			LEFT JOIN "order" o ON o.id = op.order_id
			GROUP BY p.id
		) p
		JOIN category c ON c.id = p.category_id
		JOIN seller s ON s.user_id = p.seller_id
		LEFT JOIN LATERAL (
			SELECT COALESCE(AVG(pr.rating), 0)::double precision AS rating, COUNT(*) AS review_count
			FROM product_review pr WHERE pr.product_id = p.id
		) reviews ON TRUE
		WHERE ` + strings.Join(conditions, " AND ")
	args = append(args, params.Limit+1)

	query += " ORDER BY " + SortingBySortBy(params.SortBy)
	query += fmt.Sprintf(" LIMIT $%d", len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return domain.ProductsPage{}, fmt.Errorf("search products: %w", err)
	}
	defer rows.Close()

	page := domain.ProductsPage{
		Products:     make([]domain.Product, 0, params.Limit),
		Sellers:      make([]sellerDomain.Seller, 0, params.Limit),
		PictureURLs:  make([][]string, 0, params.Limit),
		Ratings:      make([]float64, 0, params.Limit),
		ReviewCounts: make([]uint, 0, params.Limit),
	}
	for rows.Next() {
		var (
			product      domain.Product
			seller       sellerDomain.Seller
			picturesJSON []byte
			rating       float64
			reviewCount  uint
			buyCount     uint
		)
		if err := rows.Scan(
			&product.ID, &product.CategoryID, &product.SellerID, &product.Price,
			&product.Discount, &product.Name, &product.Description, &product.AvailableCount,
			&buyCount, &product.CreatedAt, &product.UpdatedAt,
			&picturesJSON, &rating, &reviewCount, &seller.UserID, &seller.Name, &seller.AvatarURL,
		); err != nil {
			return domain.ProductsPage{}, fmt.Errorf("scan product: %w", err)
		}
		if len(page.Products) < int(params.Limit) {
			var pictureURLs []string
			if err := json.Unmarshal(picturesJSON, &pictureURLs); err != nil {
				return domain.ProductsPage{}, fmt.Errorf("decode product pictures: %w", err)
			}
			page.Products = append(page.Products, product)
			page.Sellers = append(page.Sellers, seller)
			page.PictureURLs = append(page.PictureURLs, pictureURLs)
			page.Ratings = append(page.Ratings, rating)
			page.ReviewCounts = append(page.ReviewCounts, reviewCount)
			page.BuyCounts = append(page.BuyCounts, buyCount)
		} else {
			lastProduct := page.Products[len(page.Products)-1]
			page.NextCursor = &domain.Cursor{
				CreatedAt: lastProduct.CreatedAt, BuyCount: page.BuyCounts[len(page.BuyCounts)-1], Price: lastProduct.Price,
				ID: lastProduct.ID, Rating: page.Ratings[len(page.Ratings)-1],
			}
		}
	}
	if err := rows.Err(); err != nil {
		return domain.ProductsPage{}, fmt.Errorf("read products: %w", err)
	}
	return page, nil
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
