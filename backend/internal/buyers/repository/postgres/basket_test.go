package postgres

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"vibe_market/backend/internal/buyers/domain"
)

func newBasketRepoTest(t *testing.T) (*BasketRepoPostgres, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		_ = db.Close()
	})
	return NewBasketRepo(db), mock
}

func TestGetBasket(t *testing.T) {
	t.Run("buyer not found", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM buyer WHERE user_id = $1;")).
			WithArgs(int64(10)).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetBasket(10)
		if !errors.Is(err, domain.ErrBuyerNotFound) {
			t.Fatalf("got err %v, want %v", err, domain.ErrBuyerNotFound)
		}
	})

	t.Run("success returns items and sum", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM buyer WHERE user_id = $1;")).
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO basket (buyer_id) VALUES ($1) ON CONFLICT (buyer_id) DO NOTHING;")).
			WithArgs(int64(10)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		basketRows := sqlmock.NewRows([]string{
			"product_id", "category", "name", "picture_url", "price",
			"discount", "count", "available_count", "seller_name", "category_id",
		}).AddRow(1, "Electronics", "Phone", "http://img.jpg", 100.0, 10, 2, 5, "Store", 3)

		mock.ExpectQuery("SELECT bp\\.product_id,.*FROM basket b.*WHERE b\\.buyer_id = \\$1").
			WithArgs(int64(10)).
			WillReturnRows(basketRows)

		summary, err := repo.GetBasket(10)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(summary.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(summary.Items))
		}
		item := summary.Items[0]
		if item.ProductID != 1 || item.Name != "Phone" || item.Price != 100 || item.PriceResult != 90 {
			t.Errorf("unexpected item: %+v", item)
		}
		if summary.BasketSum != 200 || summary.BasketSumResult != 180 {
			t.Errorf("unexpected sums: sum=%f, sumResult=%f", summary.BasketSum, summary.BasketSumResult)
		}
	})
}

func TestAddProduct(t *testing.T) {
	t.Run("invalid count", func(t *testing.T) {
		repo, _ := newBasketRepoTest(t)
		err := repo.AddProduct(10, 1, 0)
		if !errors.Is(err, domain.ErrCountInvalid) {
			t.Fatalf("got err %v, want %v", err, domain.ErrCountInvalid)
		}
	})

	t.Run("product not found", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM buyer WHERE user_id = $1;")).
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT available_count FROM product WHERE id = $1;")).
			WithArgs(int64(999)).
			WillReturnError(sql.ErrNoRows)

		err := repo.AddProduct(10, 999, 1)
		if !errors.Is(err, domain.ErrProductNotFound) {
			t.Fatalf("got err %v, want %v", err, domain.ErrProductNotFound)
		}
	})

	t.Run("count exceeded", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM buyer WHERE user_id = $1;")).
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT available_count FROM product WHERE id = $1;")).
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"available_count"}).AddRow(5))

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO basket (buyer_id) VALUES ($1) ON CONFLICT (buyer_id) DO NOTHING;")).
			WithArgs(int64(10)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery("SELECT bp\\.count FROM basket b JOIN basket_product bp").
			WithArgs(int64(10), int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4)) // 4 already in basket

		err := repo.AddProduct(10, 1, 2) // 4 + 2 = 6 > 5
		if !errors.Is(err, domain.ErrCountExceeded) {
			t.Fatalf("got err %v, want %v", err, domain.ErrCountExceeded)
		}
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM buyer WHERE user_id = $1;")).
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT available_count FROM product WHERE id = $1;")).
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"available_count"}).AddRow(10))

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO basket (buyer_id) VALUES ($1) ON CONFLICT (buyer_id) DO NOTHING;")).
			WithArgs(int64(10)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery("SELECT bp\\.count FROM basket b JOIN basket_product bp").
			WithArgs(int64(10), int64(1)).
			WillReturnError(sql.ErrNoRows) // Not in basket yet

		mock.ExpectExec("INSERT INTO basket_product \\(basket_id, product_id, count\\)").
			WithArgs(int64(10), int64(1), 2).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.AddProduct(10, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestUpdateProductCount(t *testing.T) {
	t.Run("invalid count", func(t *testing.T) {
		repo, _ := newBasketRepoTest(t)
		err := repo.UpdateProductCount(10, 1, -1)
		if !errors.Is(err, domain.ErrCountInvalid) {
			t.Fatalf("got err %v, want %v", err, domain.ErrCountInvalid)
		}
	})

	t.Run("not in basket", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM buyer WHERE user_id = $1;")).
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		mock.ExpectQuery("SELECT p\\.available_count FROM basket b JOIN basket_product bp").
			WithArgs(int64(10), int64(1)).
			WillReturnError(sql.ErrNoRows)

		err := repo.UpdateProductCount(10, 1, 2)
		if !errors.Is(err, domain.ErrProductNotFoundInBasket) {
			t.Fatalf("got err %v, want %v", err, domain.ErrProductNotFoundInBasket)
		}
	})

	t.Run("count exceeded", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM buyer WHERE user_id = $1;")).
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		mock.ExpectQuery("SELECT p\\.available_count FROM basket b JOIN basket_product bp").
			WithArgs(int64(10), int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"available_count"}).AddRow(3))

		err := repo.UpdateProductCount(10, 1, 5) // 5 > 3
		if !errors.Is(err, domain.ErrCountExceeded) {
			t.Fatalf("got err %v, want %v", err, domain.ErrCountExceeded)
		}
	})

	t.Run("count zero deletes item", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM buyer WHERE user_id = $1;")).
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		mock.ExpectQuery("SELECT p\\.available_count FROM basket b JOIN basket_product bp").
			WithArgs(int64(10), int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"available_count"}).AddRow(10))

		mock.ExpectExec("DELETE FROM basket_product WHERE basket_id = \\(SELECT id FROM basket WHERE buyer_id = \\$1\\) AND product_id = \\$2;").
			WithArgs(int64(10), int64(1)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateProductCount(10, 1, 0)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	})

	t.Run("count positive updates item", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM buyer WHERE user_id = $1;")).
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		mock.ExpectQuery("SELECT p\\.available_count FROM basket b JOIN basket_product bp").
			WithArgs(int64(10), int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"available_count"}).AddRow(10))

		mock.ExpectExec("UPDATE basket_product SET count = \\$3, updated_at = NOW\\(\\) WHERE basket_id = \\(SELECT id FROM basket WHERE buyer_id = \\$1\\) AND product_id = \\$2;").
			WithArgs(int64(10), int64(1), 4).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateProductCount(10, 1, 4)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestGetNearestPickupPoint(t *testing.T) {
	t.Run("none found", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		mock.ExpectQuery("SELECT user_id, longitude, latitude, start_time::text, end_time::text FROM pickup_point").
			WithArgs(55.7, 37.6).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetNearestPickupPoint(55.7, 37.6)
		if !errors.Is(err, domain.ErrNoPickupPoints) {
			t.Fatalf("got err %v, want %v", err, domain.ErrNoPickupPoints)
		}
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		mock.ExpectQuery("SELECT user_id, longitude, latitude, start_time::text, end_time::text FROM pickup_point").
			WithArgs(55.7, 37.6).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "longitude", "latitude", "start_time", "end_time"}).
				AddRow(1, 37.6, 55.7, "09:00:00", "21:00:00"))

		p, err := repo.GetNearestPickupPoint(55.7, 37.6)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.UserID != 1 || p.Longitude != 37.6 || p.Latitude != 55.7 {
			t.Errorf("unexpected pickup point: %+v", p)
		}
	})
}

func TestGetPromocode(t *testing.T) {
	t.Run("expired or not found", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		mock.ExpectQuery("SELECT id, code, available_category_id, discount, start_datetime, end_datetime FROM promocode").
			WithArgs("INVALID").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetPromocode("INVALID")
		if !errors.Is(err, domain.ErrPromocodeNotFoundOrExpired) {
			t.Fatalf("got err %v, want %v", err, domain.ErrPromocodeNotFoundOrExpired)
		}
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := newBasketRepoTest(t)
		now := time.Now()
		mock.ExpectQuery("SELECT id, code, available_category_id, discount, start_datetime, end_datetime FROM promocode").
			WithArgs("SALE").
			WillReturnRows(sqlmock.NewRows([]string{"id", "code", "available_category_id", "discount", "start_datetime", "end_datetime"}).
				AddRow(5, "SALE", int64(2), 15, now.Add(-time.Hour), now.Add(time.Hour)))

		promo, err := repo.GetPromocode("SALE")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if promo.ID != 5 || promo.Code != "SALE" || promo.Discount != 15 || *promo.AvailableCategoryID != 2 {
			t.Errorf("unexpected promo: %+v", promo)
		}
	})
}
