package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"vibe_market/backend/internal/buyers/domain"
)

type BasketRepoPostgres struct {
	db *sql.DB
}

func NewBasketRepo(db *sql.DB) *BasketRepoPostgres {
	return &BasketRepoPostgres{db: db}
}

func (r *BasketRepoPostgres) ensureBuyerExists(buyerID int64) error {
	var dummy int
	err := r.db.QueryRow("SELECT 1 FROM buyer WHERE user_id = $1;", buyerID).Scan(&dummy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrBuyerNotFound
		}
		return fmt.Errorf("check buyer: %w", err)
	}
	return nil
}

func (r *BasketRepoPostgres) ensureBasketExists(buyerID int64) error {
	_, err := r.db.Exec("INSERT INTO basket (buyer_id) VALUES ($1) ON CONFLICT (buyer_id) DO NOTHING;", buyerID)
	if err != nil {
		return fmt.Errorf("ensure basket: %w", err)
	}
	return nil
}

func (r *BasketRepoPostgres) GetBasket(buyerID int64) (domain.BasketSummary, error) {
	if err := r.ensureBuyerExists(buyerID); err != nil {
		return domain.BasketSummary{}, err
	}
	if err := r.ensureBasketExists(buyerID); err != nil {
		return domain.BasketSummary{}, err
	}

	query := `
		SELECT 
			bp.product_id,
			c.category,
			p.name,
			COALESCE((SELECT pp.picture_url FROM product_picture pp WHERE pp.product_id = p.id ORDER BY pp.created_at, pp.id LIMIT 1), ''),
			p.price,
			p.discount,
			bp.count,
			p.available_count,
			s.name,
			p.category_id
		FROM basket b
		JOIN basket_product bp ON bp.basket_id = b.id
		JOIN product p ON p.id = bp.product_id
		JOIN product_category c ON c.id = p.category_id
		JOIN seller s ON s.user_id = p.seller_id
		WHERE b.buyer_id = $1
		ORDER BY bp.created_at, bp.product_id;
	`

	rows, err := r.db.Query(query, buyerID)
	if err != nil {
		return domain.BasketSummary{}, fmt.Errorf("query basket: %w", err)
	}
	defer rows.Close()

	summary := domain.BasketSummary{
		Items:           make([]domain.BasketItem, 0),
		BasketSum:       0,
		BasketSumResult: 0,
	}

	for rows.Next() {
		var item domain.BasketItem
		if err := rows.Scan(
			&item.ProductID,
			&item.CategoryName,
			&item.Name,
			&item.PictureURL,
			&item.Price,
			&item.Discount,
			&item.Count,
			&item.AvailableCount,
			&item.SellerName,
			&item.CategoryID,
		); err != nil {
			return domain.BasketSummary{}, fmt.Errorf("scan basket item: %w", err)
		}

		item.PriceResult = (1.0 - float64(item.Discount)/100.0) * item.Price
		summary.BasketSum += item.Price * float64(item.Count)
		summary.BasketSumResult += item.PriceResult * float64(item.Count)
		summary.Items = append(summary.Items, item)
	}

	if err := rows.Err(); err != nil {
		return domain.BasketSummary{}, fmt.Errorf("rows basket: %w", err)
	}

	return summary, nil
}

func (r *BasketRepoPostgres) AddProduct(buyerID int64, productID int64, count int) error {
	if count <= 0 {
		return domain.ErrCountInvalid
	}
	if err := r.ensureBuyerExists(buyerID); err != nil {
		return err
	}

	var availableCount int
	err := r.db.QueryRow("SELECT available_count FROM product WHERE id = $1;", productID).Scan(&availableCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrProductNotFound
		}
		return fmt.Errorf("check product: %w", err)
	}

	if err := r.ensureBasketExists(buyerID); err != nil {
		return err
	}

	var currentCount int
	err = r.db.QueryRow(`
		SELECT bp.count
		FROM basket b
		JOIN basket_product bp ON bp.basket_id = b.id
		WHERE b.buyer_id = $1 AND bp.product_id = $2;
	`, buyerID, productID).Scan(&currentCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check current count: %w", err)
	}

	if currentCount+count > availableCount {
		return domain.ErrCountExceeded
	}

	query := `
		INSERT INTO basket_product (basket_id, product_id, count)
		SELECT b.id, $2, $3
		FROM basket b
		WHERE b.buyer_id = $1
		ON CONFLICT (basket_id, product_id)
		DO UPDATE SET count = basket_product.count + EXCLUDED.count, updated_at = NOW();
	`
	_, err = r.db.Exec(query, buyerID, productID, count)
	if err != nil {
		return fmt.Errorf("upsert basket product: %w", err)
	}

	return nil
}

func (r *BasketRepoPostgres) UpdateProductCount(buyerID int64, productID int64, count int) error {
	if count < 0 {
		return domain.ErrCountInvalid
	}
	if err := r.ensureBuyerExists(buyerID); err != nil {
		return err
	}

	var availableCount int
	err := r.db.QueryRow(`
		SELECT p.available_count
		FROM basket b
		JOIN basket_product bp ON bp.basket_id = b.id
		JOIN product p ON p.id = bp.product_id
		WHERE b.buyer_id = $1 AND bp.product_id = $2;
	`, buyerID, productID).Scan(&availableCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrProductNotFoundInBasket
		}
		return fmt.Errorf("check product in basket: %w", err)
	}

	if count > availableCount {
		return domain.ErrCountExceeded
	}

	if count == 0 {
		_, err = r.db.Exec(`
			DELETE FROM basket_product
			WHERE basket_id = (SELECT id FROM basket WHERE buyer_id = $1)
			  AND product_id = $2;
		`, buyerID, productID)
		if err != nil {
			return fmt.Errorf("delete basket product: %w", err)
		}
		return nil
	}

	_, err = r.db.Exec(`
		UPDATE basket_product
		SET count = $3, updated_at = NOW()
		WHERE basket_id = (SELECT id FROM basket WHERE buyer_id = $1)
		  AND product_id = $2;
	`, buyerID, productID, count)
	if err != nil {
		return fmt.Errorf("update basket product: %w", err)
	}

	return nil
}

func (r *BasketRepoPostgres) GetNearestPickupPoint(lat, lng float64) (domain.PickupPoint, error) {
	query := `
		SELECT user_id, longitude, latitude, start_time::text, end_time::text
		FROM pickup_point
		ORDER BY earth_distance(ll_to_earth(latitude, longitude), ll_to_earth($1, $2)) ASC
		LIMIT 1;
	`
	var p domain.PickupPoint
	err := r.db.QueryRow(query, lat, lng).Scan(&p.UserID, &p.Longitude, &p.Latitude, &p.StartTime, &p.EndTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.PickupPoint{}, domain.ErrNoPickupPoints
		}
		return domain.PickupPoint{}, fmt.Errorf("nearest pickup point: %w", err)
	}
	return p, nil
}

func (r *BasketRepoPostgres) GetPromocode(code string) (*domain.Promocode, error) {
	query := `
		SELECT id, code, available_category_id, discount, start_datetime, end_datetime
		FROM promocode
		WHERE UPPER(code) = UPPER($1) AND NOW() BETWEEN start_datetime AND end_datetime;
	`
	var p domain.Promocode
	var catID sql.NullInt64
	err := r.db.QueryRow(query, code).Scan(&p.ID, &p.Code, &catID, &p.Discount, &p.StartDatetime, &p.EndDatetime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPromocodeNotFoundOrExpired
		}
		return nil, fmt.Errorf("get promocode: %w", err)
	}
	if catID.Valid {
		p.AvailableCategoryID = &catID.Int64
	}
	return &p, nil
}
