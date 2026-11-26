package postgres

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"vibe_market/backend/internal/products/domain"
	sellerDomain "vibe_market/backend/internal/sellers/domain"
)

func isPriceInvalid(price float64) bool {
	return price < 0 || math.IsNaN(price) || math.IsInf(price, 0)
}

func (r *ProductRepoPostgres) SearchByQuery(query string, limit uint) ([]domain.ProductSearchResult, error) {
	rows, err := r.db.Query(`
		(SELECT p.id, p.name,
			COALESCE((SELECT pp.picture_url FROM product_picture pp
				WHERE pp.product_id = p.id ORDER BY pp.created_at, pp.id LIMIT 1), ''),
			s.name
		FROM product p
		JOIN seller s ON s.user_id = p.seller_id
		WHERE p.fts @@ websearch_to_tsquery('ru', $1) AND p.available_count > 0 
		ORDER BY ts_rank(p.fts, websearch_to_tsquery('ru', $1)) DESC, p.id DESC
		LIMIT $2)
		UNION
		(SELECT p.id, p.name,
			COALESCE((SELECT pp.picture_url FROM product_picture pp
				WHERE pp.product_id = p.id ORDER BY pp.created_at, pp.id LIMIT 1), ''),
			s.name
		FROM product p
		JOIN seller s ON s.user_id = p.seller_id
		WHERE p.name ILIKE '%' || $1 || '%' AND p.available_count > 0 
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $2)
		`, query, limit)
	if err != nil {
		return []domain.ProductSearchResult{}, fmt.Errorf("search products by query: %w", err)
	}
	defer rows.Close()

	products := make([]domain.ProductSearchResult, 0, limit)
	for rows.Next() {
		var product domain.ProductSearchResult
		if err := rows.Scan(&product.ID, &product.Name, &product.PictureURL, &product.SellerName); err != nil {
			return []domain.ProductSearchResult{}, fmt.Errorf("scan product search result: %w", err)
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return []domain.ProductSearchResult{}, fmt.Errorf("read product search results: %w", err)
	}
	return products, nil
}

func addLastViewedProductByCursor(cursor domain.Cursor, sortBy domain.SortByType, args []any, conditions []string) ([]any, []string) {
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
	case domain.SortByDescRating:
		args = append(args, cursor.Rating, cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(reviews.rating, p.id) < ($%d, $%d)", len(args)-1, len(args)))
	case domain.SortByAscRating:
		args = append(args, cursor.Rating, cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(reviews.rating, p.id) > ($%d, $%d)", len(args)-1, len(args)))
	}
	return args, conditions
}

func sortingBySortBy(sortBy domain.SortByType) string {
	switch sortBy {
	case domain.SortByNew:
		return "p.created_at DESC, p.id DESC"
	case domain.SortByAscPrice:
		return "p.price ASC, p.id DESC"
	case domain.SortByDescPrice:
		return "p.price DESC, p.id DESC"
	case domain.SortByBuyCount:
		return "p.buy_count DESC, p.id DESC"
	case domain.SortByDescRating:
		return "reviews.rating DESC, p.id DESC"
	case domain.SortByAscRating:
		return "reviews.rating ASC, p.id DESC"
	default:
		return ""
	}
}

func (r *ProductRepoPostgres) Search(params domain.ProductsSearchParams) (domain.ProductsPage, error) {
	switch params.SortBy {
	case domain.SortByNew, domain.SortByAscPrice, domain.SortByDescPrice,
		domain.SortByBuyCount, domain.SortByDescRating, domain.SortByAscRating:
	default:
		return domain.ProductsPage{}, domain.ErrSortByInvalid
	}
	if isPriceInvalid(params.PriceMin) || params.PriceMax != nil &&
		(isPriceInvalid(*params.PriceMax) || params.PriceMin > *params.PriceMax) {
		return domain.ProductsPage{}, domain.ErrPriceRangeInvalid
	}
	hasCursor := params.Cursor != nil
	if hasCursor && params.SortBy == domain.SortByNew && params.Cursor.CreatedAt.IsZero() {
		return domain.ProductsPage{}, domain.ErrCursorInvalid
	}

	conditions := make([]string, 0, 8)
	args := make([]any, 0, 10)
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
	addCondition("p.price::BIGINT * (100 - p.discount) / 100 >= $%d", params.PriceMin)
	if params.PriceMax != nil {
		addCondition("p.price::BIGINT * (100 - p.discount) / 100 <= $%d", *params.PriceMax)
	}
	addCondition("p.available_count > $%d", 0)

	priceQuery := `SELECT COALESCE(MIN(p.price::BIGINT * (100 - p.discount) / 100), 0),
		COALESCE(MAX(p.price::BIGINT * (100 - p.discount) / 100), 0)
		FROM product p
		JOIN category c ON c.id = p.category_id
		JOIN seller s ON s.user_id = p.seller_id
		WHERE ` + strings.Join(conditions, " AND ")
	var priceResultMin, priceResultMax int
	if err := r.db.QueryRow(priceQuery, args...).Scan(&priceResultMin, &priceResultMax); err != nil {
		return domain.ProductsPage{}, fmt.Errorf("search product price range: %w", err)
	}

	if hasCursor {
		args, conditions = addLastViewedProductByCursor(*params.Cursor, params.SortBy, args, conditions)
	}
	buyCountColumn := ""
	productSource := "product p"
	if params.SortBy == domain.SortByBuyCount {
		buyCountColumn = "p.buy_count,"
		productSource = `(SELECT p.*, COALESCE(SUM(op.count) FILTER (WHERE o.order_status_id = 4), 0) AS buy_count
			FROM product p
			LEFT JOIN order_product op ON op.product_id = p.id
			LEFT JOIN "order" o ON o.id = op.order_id
			GROUP BY p.id) p`
	}
	query := `SELECT p.id, p.category_id, p.seller_id, p.price, p.discount, p.name,
		p.description, p.available_count, ` + buyCountColumn + ` p.created_at, p.updated_at,
		COALESCE((SELECT json_agg(pp.picture_url ORDER BY pp.created_at, pp.id)
			FROM product_picture pp WHERE pp.product_id = p.id), '[]'::json),
		reviews.rating, reviews.review_count, s.user_id, s.name, s.avatar_url
		FROM ` + productSource + `
		JOIN category c ON c.id = p.category_id
		JOIN seller s ON s.user_id = p.seller_id
		LEFT JOIN LATERAL (
			SELECT COALESCE(AVG(pr.rating), 0)::double precision AS rating, COUNT(*) AS review_count
			FROM product_review pr WHERE pr.product_id = p.id
		) reviews ON TRUE
		WHERE ` + strings.Join(conditions, " AND ")
	args = append(args, params.Limit+1)

	query += " ORDER BY " + sortingBySortBy(params.SortBy)
	query += fmt.Sprintf(" LIMIT $%d", len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return domain.ProductsPage{}, fmt.Errorf("search products: %w", err)
	}
	defer rows.Close()

	page := domain.ProductsPage{
		Products:       make([]domain.Product, 0, params.Limit),
		Sellers:        make([]sellerDomain.Seller, 0, params.Limit),
		PictureURLs:    make([][]string, 0, params.Limit),
		Ratings:        make([]float64, 0, params.Limit),
		ReviewCounts:   make([]uint, 0, params.Limit),
		PriceResultMin: priceResultMin,
		PriceResultMax: priceResultMax,
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
		scanArgs := []any{
			&product.ID, &product.CategoryID, &product.SellerID, &product.Price,
			&product.Discount, &product.Name, &product.Description, &product.AvailableCount,
		}
		if params.SortBy == domain.SortByBuyCount {
			scanArgs = append(scanArgs, &buyCount)
		}
		scanArgs = append(scanArgs,
			&product.CreatedAt, &product.UpdatedAt, &picturesJSON, &rating, &reviewCount,
			&seller.UserID, &seller.Name, &seller.AvatarURL,
		)
		if err := rows.Scan(scanArgs...); err != nil {
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
			page.NextCursor = &domain.Cursor{ID: lastProduct.ID}
			switch params.SortBy {
			case domain.SortByNew:
				page.NextCursor.CreatedAt = lastProduct.CreatedAt
			case domain.SortByBuyCount:
				page.NextCursor.BuyCount = page.BuyCounts[len(page.BuyCounts)-1]
			case domain.SortByAscPrice, domain.SortByDescPrice:
				page.NextCursor.Price = lastProduct.Price
			case domain.SortByDescRating, domain.SortByAscRating:
				page.NextCursor.Rating = page.Ratings[len(page.Ratings)-1]
			}
		}
	}
	if err := rows.Err(); err != nil {
		return domain.ProductsPage{}, fmt.Errorf("read products: %w", err)
	}
	return page, nil
}
