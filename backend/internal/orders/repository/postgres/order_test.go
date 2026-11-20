package postgres

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"vibe_market/backend/internal/orders/domain"
)

const (
	orderTestBuyerID       int64 = 11
	orderTestPickupPointID int64 = 73
	orderTestOrderID       int64 = 101

	orderTestPickupQuery = `SELECT user_id FROM pickup_point WHERE user_id = $1 FOR KEY SHARE;`
	orderTestPromoQuery  = `SELECT id, available_category_id, discount FROM promocode
		WHERE UPPER(code) = UPPER($1) AND NOW() BETWEEN start_datetime AND end_datetime;`
)

func newOrderRepoTest(t *testing.T) (*OrderRepoPostgres, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		_ = db.Close()
	})
	return NewOrderRepo(db), mock
}

func TestSelectByProductIDAndBuyerID(t *testing.T) {
	const query = `SELECT o.id, o.buyer_id, o.order_status_id, o.promocode_id,
		o.pickup_point_id, o.created_at, o.updated_at FROM "order" o
		WHERE o.buyer_id = $2 AND EXISTS ( SELECT 1 FROM order_product op
		WHERE op.order_id = o.id AND op.product_id = $1 );`
	columns := []string{"id", "buyer_id", "order_status_id", "promocode_id", "pickup_point_id", "created_at", "updated_at"}
	now := time.Now()
	t.Run("multiple orders", func(t *testing.T) {
		repo, mock := newOrderRepoTest(t)
		mock.ExpectQuery(query).WithArgs(int64(21), orderTestBuyerID).
			WillReturnRows(sqlmock.NewRows(columns).
				AddRow(int64(101), orderTestBuyerID, int64(1), nil, orderTestPickupPointID, now, now).
				AddRow(int64(102), orderTestBuyerID, int64(2), int64(41), orderTestPickupPointID, now, now)).RowsWillBeClosed()
		orders, err := repo.SelectByProductIDAndBuyerID(21, orderTestBuyerID)
		if err != nil || len(orders) != 2 {
			t.Fatalf("orders=%+v, err=%v", orders, err)
		}
		if orders[0].ID != 101 || orders[0].PromocodeID != nil || orders[1].ID != 102 || orders[1].PromocodeID == nil || *orders[1].PromocodeID != 41 {
			t.Fatalf("unexpected orders: %+v", orders)
		}
	})
	t.Run("no orders", func(t *testing.T) {
		repo, mock := newOrderRepoTest(t)
		mock.ExpectQuery(query).WithArgs(int64(21), orderTestBuyerID).WillReturnRows(sqlmock.NewRows(columns)).RowsWillBeClosed()
		orders, err := repo.SelectByProductIDAndBuyerID(21, orderTestBuyerID)
		if err != nil || len(orders) != 0 {
			t.Fatalf("orders=%+v, err=%v", orders, err)
		}
	})
	t.Run("query error", func(t *testing.T) {
		repo, mock := newOrderRepoTest(t)
		dbErr := errors.New("database unavailable")
		mock.ExpectQuery(query).WithArgs(int64(21), orderTestBuyerID).WillReturnError(dbErr)
		_, err := repo.SelectByProductIDAndBuyerID(21, orderTestBuyerID)
		if !errors.Is(err, dbErr) {
			t.Fatalf("got error %v, want %v", err, dbErr)
		}
	})
}

func expectOrderBasket(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT bp.product_id, bp.count, p.name, p.price, p.discount,
		p.available_count, s.name, p.category_id
		FROM basket b
		JOIN basket_product bp ON bp.basket_id = b.id
		JOIN product p ON p.id = bp.product_id
		JOIN seller s ON s.user_id = p.seller_id
		WHERE b.buyer_id = $1 FOR UPDATE OF p;`).
		WithArgs(orderTestBuyerID).
		WillReturnRows(sqlmock.NewRows([]string{
			"product_id", "count", "name", "price", "discount", "available_count", "seller_name", "category_id",
		}).AddRow(int64(21), 2, "Product", 100.0, 10, 5, "Seller", int64(31))).
		RowsWillBeClosed()
}

func TestCreateOrderTxSuccess(t *testing.T) {
	tests := []struct {
		name          string
		promo         string
		promoID       driver.Value
		promoCategory driver.Value
		wantPrice     int
	}{
		{name: "without promo", wantPrice: 90},
		{name: "whitespace promo", promo: " \t ", wantPrice: 90},
		{name: "global promo", promo: "  SAVE20  ", promoID: int64(41), wantPrice: 72},
		{name: "matching category promo", promo: "SAVE20", promoID: int64(41), promoCategory: int64(31), wantPrice: 72},
		{name: "other category promo", promo: "SAVE20", promoID: int64(41), promoCategory: int64(32), wantPrice: 90},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := newOrderRepoTest(t)
			expectOrderBasket(mock)
			mock.ExpectQuery(orderTestPickupQuery).
				WithArgs(orderTestPickupPointID).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(orderTestPickupPointID))
			if test.promoID != nil {
				mock.ExpectQuery(orderTestPromoQuery).
					WithArgs("SAVE20").
					WillReturnRows(sqlmock.NewRows([]string{"id", "available_category_id", "discount"}).
						AddRow(test.promoID, test.promoCategory, 20))
			}
			mock.ExpectQuery(`INSERT INTO "order" (buyer_id, order_status_id, promocode_id, pickup_point_id)
				VALUES ($1, 1, $2, $3) RETURNING id;`).
				WithArgs(orderTestBuyerID, test.promoID, orderTestPickupPointID).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(orderTestOrderID))
			mock.ExpectExec(`INSERT INTO order_product (order_id, product_id, count, price)
				VALUES ($1, $2, $3, $4);`).
				WithArgs(orderTestOrderID, int64(21), 2, test.wantPrice).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec(`UPDATE product SET available_count = available_count - $2, updated_at = NOW()
				WHERE id = $1;`).
				WithArgs(int64(21), 2).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec(`DELETE FROM basket_product
				WHERE basket_id = (SELECT id FROM basket WHERE buyer_id = $1);`).
				WithArgs(orderTestBuyerID).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			orderID, stockErrors, err := repo.CreateOrderTx(context.Background(), orderTestBuyerID, orderTestPickupPointID, test.promo)
			if err != nil {
				t.Fatal(err)
			}
			if orderID != orderTestOrderID || stockErrors != nil {
				t.Fatalf("unexpected result: orderID=%d, stockErrors=%+v", orderID, stockErrors)
			}
		})
	}
}

func TestCreateOrderTxPickupPointFailure(t *testing.T) {
	dbErr := errors.New("pickup point database unavailable")
	tests := []struct {
		name  string
		dbErr error
		want  error
	}{
		{name: "not found", want: domain.ErrPickupPointInvalid},
		{name: "database error", dbErr: dbErr, want: dbErr},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := newOrderRepoTest(t)
			expectOrderBasket(mock)
			query := mock.ExpectQuery(orderTestPickupQuery).WithArgs(orderTestPickupPointID)
			if test.dbErr != nil {
				query.WillReturnError(test.dbErr)
			} else {
				query.WillReturnRows(sqlmock.NewRows([]string{"user_id"}))
			}
			mock.ExpectRollback()

			orderID, stockErrors, err := repo.CreateOrderTx(context.Background(), orderTestBuyerID, orderTestPickupPointID, "SAVE20")
			if !errors.Is(err, test.want) {
				t.Fatalf("got error %v, want %v", err, test.want)
			}
			if orderID != 0 || stockErrors != nil {
				t.Fatalf("unexpected result: orderID=%d, stockErrors=%+v", orderID, stockErrors)
			}
		})
	}
}

func TestCreateOrderTxInvalidPromoRollsBack(t *testing.T) {
	repo, mock := newOrderRepoTest(t)
	expectOrderBasket(mock)
	mock.ExpectQuery(orderTestPickupQuery).
		WithArgs(orderTestPickupPointID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(orderTestPickupPointID))
	mock.ExpectQuery(orderTestPromoQuery).
		WithArgs("EXPIRED").
		WillReturnRows(sqlmock.NewRows([]string{"id", "available_category_id", "discount"}))
	mock.ExpectRollback()

	orderID, stockErrors, err := repo.CreateOrderTx(context.Background(), orderTestBuyerID, orderTestPickupPointID, " EXPIRED ")
	if !errors.Is(err, domain.ErrPromocodeInvalid) {
		t.Fatalf("got error %v, want %v", err, domain.ErrPromocodeInvalid)
	}
	if orderID != 0 || stockErrors != nil {
		t.Fatalf("unexpected result: orderID=%d, stockErrors=%+v", orderID, stockErrors)
	}
}
