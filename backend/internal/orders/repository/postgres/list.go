package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"vibe_market/backend/internal/orders/domain"
)

const listOrdersQuery = `
 SELECT o.id, os.id, os.status, o.created_at, o.updated_at,
 pp.user_id, pp.longitude, pp.latitude, pp.start_time::text, pp.end_time::text,
 pc.id, pc.code, pc.discount, pc.available_category_id, pc.start_datetime, pc.end_datetime
 FROM "order" o
 JOIN order_status os ON os.id = o.order_status_id
 JOIN pickup_point pp ON pp.user_id = o.pickup_point_id
 LEFT JOIN promocode pc ON pc.id = o.promocode_id
 WHERE o.buyer_id = $1
 ORDER BY o.created_at DESC, o.id DESC`

const listProductsQuery = `
 SELECT op.order_id, op.id, p.id, p.name, p.description,
 c.id, c.category, s.user_id, s.name, COALESCE(s.avatar_url, ''), op.count, op.price
 FROM order_product op
 JOIN "order" o ON o.id = op.order_id
 JOIN product p ON p.id = op.product_id
 JOIN category c ON c.id = p.category_id
 JOIN seller s ON s.user_id = p.seller_id
 WHERE o.buyer_id = $1
 ORDER BY op.order_id, op.id`

const listPicturesQuery = `
 SELECT op.order_id, op.id, pp.picture_url
 FROM product_picture pp
 JOIN order_product op ON op.product_id = pp.product_id
 JOIN "order" o ON o.id = op.order_id
 WHERE o.buyer_id = $1
 ORDER BY op.order_id, op.id, pp.id`

// ListByBuyer reads all collections from the same snapshot without joining
// independent one-to-many relationships or issuing per-order queries.
func (r *OrderRepoPostgres) ListByBuyer(ctx context.Context, buyerID int64) ([]domain.OrderDetails, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin list orders tx: %w", err)
	}
	defer tx.Rollback()

	orders := make([]domain.OrderDetails, 0)
	orderIndexes := make(map[int64]int)
	err = readOrderRows(ctx, tx, listOrdersQuery, buyerID, func(rows *sql.Rows) error {
		o := domain.OrderDetails{
			Products: make([]domain.OrderProductDetails, 0),
		}
		var promoID, categoryID, discount sql.NullInt64
		var code sql.NullString
		var start, end sql.NullTime
		if err := rows.Scan(&o.ID, &o.Status.ID, &o.Status.Name, &o.CreatedAt, &o.UpdatedAt,
			&o.PickupPoint.ID, &o.PickupPoint.Longitude, &o.PickupPoint.Latitude,
			&o.PickupPoint.StartTime, &o.PickupPoint.EndTime,
			&promoID, &code, &discount, &categoryID, &start, &end); err != nil {
			return err
		}
		if promoID.Valid {
			o.Promocode = &domain.OrderPromocode{ID: promoID.Int64, Code: code.String,
				Discount: int(discount.Int64), StartDatetime: start.Time, EndDatetime: end.Time}
			if categoryID.Valid {
				id := categoryID.Int64
				o.Promocode.AvailableCategoryID = &id
			}
		}
		orderIndexes[o.ID] = len(orders)
		orders = append(orders, o)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}

	if len(orders) > 0 {
		// Store indexes, not pointers into slices which may move while appending.
		productIndexes := make(map[int64]int)
		err = readOrderRows(ctx, tx, listProductsQuery, buyerID, func(rows *sql.Rows) error {
			var orderID int64
			p := domain.OrderProductDetails{PictureURLs: make([]string, 0)}
			var description sql.NullString
			if err := rows.Scan(&orderID, &p.ID, &p.ProductID, &p.Name, &description,
				&p.Category.ID, &p.Category.Name, &p.Seller.ID, &p.Seller.Name, &p.Seller.AvatarURL, &p.Count, &p.Price); err != nil {
				return err
			}
			if description.Valid {
				p.Description = &description.String
			}
			o := &orders[orderIndexes[orderID]]
			productIndexes[p.ID] = len(o.Products)
			o.Products = append(o.Products, p)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("list order products: %w", err)
		}
		err = readOrderRows(ctx, tx, listPicturesQuery, buyerID, func(rows *sql.Rows) error {
			var orderID, itemID int64
			var url string
			if err := rows.Scan(&orderID, &itemID, &url); err != nil {
				return err
			}
			p := &orders[orderIndexes[orderID]].Products[productIndexes[itemID]]
			p.PictureURLs = append(p.PictureURLs, url)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("list order pictures: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list orders tx: %w", err)
	}
	return orders, nil
}

func readOrderRows(ctx context.Context, tx *sql.Tx, query string, buyerID int64, scan func(*sql.Rows) error) error {
	rows, err := tx.QueryContext(ctx, query, buyerID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows: %w", err)
	}
	return rows.Close()
}
