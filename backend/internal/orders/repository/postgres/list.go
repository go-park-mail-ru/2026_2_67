package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"vibe_market/backend/internal/orders/domain"
)

const listOrdersQuery = `
 SELECT o.id, os.status, o.created_at, o.updated_at,
 pp.user_id, pp.longitude, pp.latitude, pp.start_time::text, pp.end_time::text,
 pc.code
 FROM "order" o
 JOIN order_status os ON os.id = o.order_status_id
 JOIN pickup_point pp ON pp.user_id = o.pickup_point_id
 LEFT JOIN promocode pc ON pc.id = o.promocode_id
	WHERE o.buyer_id = $1`

const listProductsQuery = `
 SELECT op.order_id, op.id, p.id, p.name, COALESCE(p.description, ''),
 c.category, s.user_id, s.name, s.avatar_url, op.count, op.price
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

func (r *OrderRepoPostgres) Orders(ctx context.Context, buyerID int64, params domain.OrdersParams) (domain.OrderPage, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return domain.OrderPage{}, fmt.Errorf("begin list orders tx: %w", err)
	}
	defer tx.Rollback()

	orders := make([]domain.OrderView, 0)
	orderIndexes := make(map[int64]int)
	query, args := buildOrdersQuery(params, buyerID)
	err = readOrderRows(ctx, tx, query, args, func(rows *sql.Rows) error {
		o := domain.OrderView{
			Products: make([]domain.OrderProductView, 0),
		}
		var code sql.NullString
		var startTime, endTime string
		if err := rows.Scan(&o.ID, &o.Status, &o.CreatedAt, &o.UpdatedAt,
			&o.PickupPoint.ID, &o.PickupPoint.Longitude, &o.PickupPoint.Latitude,
			&startTime, &endTime, &code); err != nil {
			return err
		}
		o.PickupPoint.StartTime = parsePickupPointTime(startTime)
		o.PickupPoint.EndTime = parsePickupPointTime(endTime)
		if code.Valid {
			o.Promocode = &code.String
		}
		orderIndexes[o.ID] = len(orders)
		orders = append(orders, o)
		return nil
	})
	if err != nil {
		return domain.OrderPage{}, fmt.Errorf("list orders: %w", err)
	}
	hasNextPage := uint(len(orders)) > params.Limit
	if hasNextPage {
		orders = orders[:params.Limit]
	}

	if len(orders) > 0 {
		// Store indexes, not pointers into slices which may move while appending.
		productIndexes := make(map[int64]int)
		err = readOrderRows(ctx, tx, listProductsQuery, []any{buyerID}, func(rows *sql.Rows) error {
			var orderID, itemID int64
			p := domain.OrderProductView{OrderProduct: domain.OrderProduct{}, PictureURLS: make([]string, 0)}
			var avatarURL *string
			if err := rows.Scan(&orderID, &itemID, &p.ProductID, &p.Name, &p.Description,
				&p.Category, &p.Seller.UserID, &p.Seller.Name, &avatarURL, &p.Count, &p.Price); err != nil {
				return err
			}
			p.OrderID = orderID
			p.Seller.AvatarURL = avatarURL
			p.TotalPrice = int(p.Count) * p.Price
			orderIndex, ok := orderIndexes[orderID]
			if !ok {
				return nil
			}
			o := &orders[orderIndex]
			productIndexes[itemID] = len(o.Products)
			o.Products = append(o.Products, p)
			return nil
		})
		if err != nil {
			return domain.OrderPage{}, fmt.Errorf("list order products: %w", err)
		}
		err = readOrderRows(ctx, tx, listPicturesQuery, []any{buyerID}, func(rows *sql.Rows) error {
			var orderID, itemID int64
			var url string
			if err := rows.Scan(&orderID, &itemID, &url); err != nil {
				return err
			}
			orderIndex, ok := orderIndexes[orderID]
			if !ok {
				return nil
			}
			productIndex, ok := productIndexes[itemID]
			if !ok {
				return nil
			}
			p := &orders[orderIndex].Products[productIndex]
			p.PictureURLS = append(p.PictureURLS, url)
			return nil
		})
		if err != nil {
			return domain.OrderPage{}, fmt.Errorf("list order pictures: %w", err)
		}
	}
	for i := range orders {
		for j := range orders[i].Products {
			orders[i].TotalCount += orders[i].Products[j].Count
			orders[i].TotalPrice += orders[i].Products[j].TotalPrice
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.OrderPage{}, fmt.Errorf("commit list orders tx: %w", err)
	}
	page := domain.OrderPage{Orders: orders}
	if hasNextPage && len(orders) > 0 {
		last := orders[len(orders)-1]
		page.NextCursor = &domain.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return page, nil
}

func parsePickupPointTime(value string) time.Time {
	parsed, _ := time.Parse("15:04:05", value)
	if parsed.IsZero() {
		parsed, _ = time.Parse("15:04", value)
	}
	return parsed
}

func readOrderRows(ctx context.Context, tx *sql.Tx, query string, args []any, scan func(*sql.Rows) error) error {
	rows, err := tx.QueryContext(ctx, query, args...)
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

func buildOrdersQuery(params domain.OrdersParams, buyerID int64) (string, []any) {
	query := listOrdersQuery
	args := []any{buyerID}
	argNumber := 2
	if params.Status != "" {
		query += fmt.Sprintf(" AND os.status = $%d", argNumber)
		args = append(args, params.Status)
		argNumber++
	}
	if params.Cursor != nil {
		operator := "<"
		if params.SortBy == domain.OrderSortByAscDate {
			operator = ">"
		}
		query += fmt.Sprintf(" AND (o.created_at, o.id) %s ($%d, $%d)", operator, argNumber, argNumber+1)
		args = append(args, params.Cursor.CreatedAt, params.Cursor.ID)
		argNumber += 2
	}
	direction := "DESC"
	if params.SortBy == domain.OrderSortByAscDate {
		direction = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY o.created_at %s, o.id %s LIMIT $%d", direction, direction, argNumber)
	args = append(args, params.Limit+1)
	return query, args
}
