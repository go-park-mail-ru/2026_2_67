package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"vibe_market/backend/internal/orders/domain"
)

type OrderRepoPostgres struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepoPostgres {
	return &OrderRepoPostgres{db: db}
}

type basketItemInternal struct {
	ProductID      int64
	Count          int
	Name           string
	Price          float64
	Discount       int
	AvailableCount int
	SellerName     string
	CategoryID     int64
}

func (r *OrderRepoPostgres) CreateOrderTx(
	ctx context.Context,
	buyerID int64,
	lat, lng float64,
	promoCode string,
) (int64, []domain.ProductStockError, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	queryItems := `
		SELECT 
			bp.product_id,
			bp.count,
			p.name,
			p.price,
			p.discount,
			p.available_count,
			s.name,
			p.category_id
		FROM basket b
		JOIN basket_product bp ON bp.basket_id = b.id
		JOIN product p ON p.id = bp.product_id
		JOIN seller s ON s.user_id = p.seller_id
		WHERE b.buyer_id = $1
		FOR UPDATE OF p;
	`

	rows, err := tx.QueryContext(ctx, queryItems, buyerID)
	if err != nil {
		return 0, nil, fmt.Errorf("select basket items for order: %w", err)
	}

	items := make([]basketItemInternal, 0)
	for rows.Next() {
		var it basketItemInternal
		if err := rows.Scan(
			&it.ProductID,
			&it.Count,
			&it.Name,
			&it.Price,
			&it.Discount,
			&it.AvailableCount,
			&it.SellerName,
			&it.CategoryID,
		); err != nil {
			rows.Close()
			return 0, nil, fmt.Errorf("scan basket item: %w", err)
		}
		items = append(items, it)
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		return 0, nil, fmt.Errorf("rows error: %w", err)
	}

	if len(items) == 0 {
		return 0, nil, domain.ErrOrderBasketEmpty
	}

	stockErrors := make([]domain.ProductStockError, 0)
	for _, it := range items {
		if it.Count > it.AvailableCount {
			stockErrors = append(stockErrors, domain.ProductStockError{
				ProductID:  it.ProductID,
				Name:       it.Name,
				Seller:     it.SellerName,
				ErrMessage: "товары закончились или их недостаточно на складе",
			})
		}
	}

	if len(stockErrors) > 0 {
		return 0, stockErrors, domain.ErrOrderStockInsufficient
	}

	var pickupPointID int64
	pickupQuery := `
		SELECT user_id
		FROM pickup_point
		ORDER BY earth_distance(ll_to_earth(latitude, longitude), ll_to_earth($1, $2)) ASC
		LIMIT 1;
	`
	err = tx.QueryRowContext(ctx, pickupQuery, lat, lng).Scan(&pickupPointID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil, domain.ErrNoPickupPointsAvailable
		}
		return 0, nil, fmt.Errorf("nearest pickup point: %w", err)
	}

	var promoID *int64
	var promoCatID sql.NullInt64
	var promoDiscount int

	cleanPromo := strings.TrimSpace(promoCode)
	if cleanPromo != "" {
		var pid int64
		promoQuery := `
			SELECT id, available_category_id, discount
			FROM promocode
			WHERE UPPER(code) = UPPER($1) AND NOW() BETWEEN start_datetime AND end_datetime;
		`
		err = tx.QueryRowContext(ctx, promoQuery, cleanPromo).Scan(&pid, &promoCatID, &promoDiscount)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, nil, domain.ErrPromocodeInvalid
			}
			return 0, nil, fmt.Errorf("check promocode: %w", err)
		}
		promoID = &pid
	}

	var orderID int64
	createOrderQuery := `
		INSERT INTO "order" (buyer_id, order_status_id, promocode_id, pickup_point_id)
		VALUES ($1, 1, $2, $3)
		RETURNING id;
	`
	err = tx.QueryRowContext(ctx, createOrderQuery, buyerID, promoID, pickupPointID).Scan(&orderID)
	if err != nil {
		return 0, nil, fmt.Errorf("insert order: %w", err)
	}

	for _, it := range items {
		discountFactor := 1.0 - float64(it.Discount)/100.0
		if promoID != nil && (!promoCatID.Valid || promoCatID.Int64 == it.CategoryID) {
			discountFactor *= 1.0 - float64(promoDiscount)/100.0
		}
		unitPrice := int(math.Round(it.Price * discountFactor))
		if unitPrice < 0 {
			unitPrice = 0
		}

		insertOrderItemQuery := `
			INSERT INTO order_product (order_id, product_id, count, price)
			VALUES ($1, $2, $3, $4);
		`
		if _, err := tx.ExecContext(ctx, insertOrderItemQuery, orderID, it.ProductID, it.Count, unitPrice); err != nil {
			return 0, nil, fmt.Errorf("insert order_product: %w", err)
		}

		decrementQuery := `
			UPDATE product
			SET available_count = available_count - $2, updated_at = NOW()
			WHERE id = $1;
		`
		if _, err := tx.ExecContext(ctx, decrementQuery, it.ProductID, it.Count); err != nil {
			return 0, nil, fmt.Errorf("decrement product stock: %w", err)
		}
	}

	insertNotificationQuery := `
		INSERT INTO order_notification (order_id, order_status_id, is_read)
		VALUES ($1, 1, FALSE);
	`
	if _, err := tx.ExecContext(ctx, insertNotificationQuery, orderID); err != nil {
		return 0, nil, fmt.Errorf("insert order_notification: %w", err)
	}

	clearBasketQuery := `
		DELETE FROM basket_product
		WHERE basket_id = (SELECT id FROM basket WHERE buyer_id = $1);
	`
	if _, err := tx.ExecContext(ctx, clearBasketQuery, buyerID); err != nil {
		return 0, nil, fmt.Errorf("clear basket: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, fmt.Errorf("commit order tx: %w", err)
	}

	return orderID, nil, nil
}
