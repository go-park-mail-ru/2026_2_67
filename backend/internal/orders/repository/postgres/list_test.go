package postgres

import (
	"context"
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"vibe_market/backend/internal/orders/domain"
)

var listTestTime = time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)

func listTestRows(stage int) *sqlmock.Rows {
	switch stage {
	case 0:
		return sqlmock.NewRows([]string{"id", "status_id", "status", "created_at", "updated_at",
			"pickup_id", "longitude", "latitude", "start_time", "end_time",
			"promo_id", "code", "discount", "category_id", "start", "end"})
	case 1:
		return sqlmock.NewRows([]string{"order_id", "id", "product_id", "name", "description",
			"category_id", "category", "seller_id", "seller", "count", "price"})
	case 2:
		return sqlmock.NewRows([]string{"order_id", "item_id", "url"})
	default:
		return sqlmock.NewRows([]string{"order_id", "id", "status_id", "status", "is_read", "created_at"})
	}
}

func listTestOrder(id int64, promoID, categoryID driver.Value) []driver.Value {
	row := []driver.Value{id, int64(1), "Processing", listTestTime, listTestTime,
		int64(73), 37.6, 55.7, "09:00:00", "21:00:00", nil, nil, nil, nil, nil, nil}
	if promoID != nil {
		row[10], row[11], row[12], row[13], row[14], row[15] = promoID, "SAVE", 20, categoryID, listTestTime, listTestTime.Add(time.Hour)
	}
	return row
}

var listTestQueries = []string{listOrdersQuery, listProductsQuery, listPicturesQuery, listNotificationsQuery}

func TestListByBuyerSuccess(t *testing.T) {
	repo, mock := newOrderRepoTest(t)
	mock.ExpectBegin()
	mock.ExpectQuery(listOrdersQuery).WithArgs(orderTestBuyerID).WillReturnRows(listTestRows(0).
		AddRow(listTestOrder(103, int64(41), int64(31))...).
		AddRow(listTestOrder(102, int64(42), nil)...).
		AddRow(listTestOrder(101, nil, nil)...)).RowsWillBeClosed()
	mock.ExpectQuery(listProductsQuery).WithArgs(orderTestBuyerID).WillReturnRows(listTestRows(1).
		AddRow(102, 202, 21, "Shared", nil, 31, "Category", 51, "Seller", 1, int64(5000000000)).
		AddRow(103, 203, 21, "Shared", "Description", 31, "Category", 51, "Seller", 2, 72).
		AddRow(103, 204, 22, "Other", nil, 32, "Other category", 52, "Other seller", 3, 0)).RowsWillBeClosed()
	mock.ExpectQuery(listPicturesQuery).WithArgs(orderTestBuyerID).WillReturnRows(listTestRows(2).
		AddRow(102, 202, "first").AddRow(102, 202, "second").
		AddRow(103, 203, "first").AddRow(103, 203, "second")).RowsWillBeClosed()
	mock.ExpectQuery(listNotificationsQuery).WithArgs(orderTestBuyerID).WillReturnRows(listTestRows(3).
		AddRow(102, 301, 1, "Processing", true, listTestTime).
		AddRow(103, 303, 2, "Shipping", false, listTestTime.Add(time.Minute)).
		AddRow(103, 302, 1, "Processing", true, listTestTime)).RowsWillBeClosed()
	mock.ExpectCommit()

	got, err := repo.ListByBuyer(context.Background(), orderTestBuyerID)
	if err != nil {
		t.Fatal(err)
	}
	categoryID := int64(31)
	description := "Description"
	base := func(id int64) domain.OrderDetails {
		return domain.OrderDetails{ID: id, Status: domain.OrderStatus{ID: 1, Name: "Processing"},
			CreatedAt: listTestTime, UpdatedAt: listTestTime,
			PickupPoint: domain.OrderPickupPoint{ID: 73, Longitude: 37.6, Latitude: 55.7, StartTime: "09:00:00", EndTime: "21:00:00"},
			Products:    []domain.OrderProductDetails{}, Notifications: []domain.OrderNotification{}}
	}
	want := []domain.OrderDetails{base(103), base(102), base(101)}
	want[0].Promocode = &domain.OrderPromocode{ID: 41, Code: "SAVE", Discount: 20, AvailableCategoryID: &categoryID,
		StartDatetime: listTestTime, EndDatetime: listTestTime.Add(time.Hour)}
	want[1].Promocode = &domain.OrderPromocode{ID: 42, Code: "SAVE", Discount: 20,
		StartDatetime: listTestTime, EndDatetime: listTestTime.Add(time.Hour)}
	product := domain.OrderProductDetails{ID: 203, ProductID: 21, Name: "Shared", Description: &description,
		Category: domain.OrderCategory{ID: 31, Name: "Category"}, Seller: domain.OrderSeller{ID: 51, Name: "Seller"},
		Count: 2, Price: 72, PictureURLs: []string{"first", "second"}}
	want[0].Products = []domain.OrderProductDetails{product, {ID: 204, ProductID: 22, Name: "Other",
		Category: domain.OrderCategory{ID: 32, Name: "Other category"}, Seller: domain.OrderSeller{ID: 52, Name: "Other seller"},
		Count: 3, Price: 0, PictureURLs: []string{}}}
	product.ID, product.Description, product.Count, product.Price = 202, nil, 1, 5000000000
	want[1].Products = []domain.OrderProductDetails{product}
	want[0].Notifications = []domain.OrderNotification{
		{ID: 303, Status: domain.OrderStatus{ID: 2, Name: "Shipping"}, CreatedAt: listTestTime.Add(time.Minute)},
		{ID: 302, Status: domain.OrderStatus{ID: 1, Name: "Processing"}, IsRead: true, CreatedAt: listTestTime}}
	want[1].Notifications = []domain.OrderNotification{{ID: 301, Status: domain.OrderStatus{ID: 1, Name: "Processing"}, IsRead: true, CreatedAt: listTestTime}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestListByBuyerEmpty(t *testing.T) {
	repo, mock := newOrderRepoTest(t)
	mock.ExpectBegin()
	mock.ExpectQuery(listOrdersQuery).WithArgs(orderTestBuyerID).WillReturnRows(listTestRows(0)).RowsWillBeClosed()
	mock.ExpectCommit()
	got, err := repo.ListByBuyer(context.Background(), orderTestBuyerID)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestListByBuyerErrors(t *testing.T) {
	failure := errors.New("database failure")
	for stage, query := range listTestQueries {
		for _, kind := range []string{"query", "scan", "rows", "close"} {
			t.Run(kind+"/"+query, func(t *testing.T) {
				repo, mock := newOrderRepoTest(t)
				mock.ExpectBegin()
				for previous := 0; previous < stage; previous++ {
					rows := listTestRows(previous)
					if previous == 0 {
						rows.AddRow(listTestOrder(101, nil, nil)...)
					}
					mock.ExpectQuery(listTestQueries[previous]).WithArgs(orderTestBuyerID).WillReturnRows(rows).RowsWillBeClosed()
				}
				expectation := mock.ExpectQuery(query).WithArgs(orderTestBuyerID)
				switch kind {
				case "query":
					expectation.WillReturnError(failure)
				case "scan":
					expectation.WillReturnRows(sqlmock.NewRows([]string{"bad"}).AddRow("bad")).RowsWillBeClosed()
				case "rows":
					expectation.WillReturnRows(listTestRows(stage).AddRow(make([]driver.Value, []int{16, 11, 3, 6}[stage])...).RowError(0, failure)).RowsWillBeClosed()
				case "close":
					expectation.WillReturnRows(listTestRows(stage).CloseError(failure)).RowsWillBeClosed()
				}
				mock.ExpectRollback()
				got, err := repo.ListByBuyer(context.Background(), orderTestBuyerID)
				if err == nil || got != nil {
					t.Fatalf("got %v, %v", got, err)
				}
				if kind == "scan" {
					if !strings.Contains(err.Error(), "scan:") {
						t.Fatalf("unexpected error: %v", err)
					}
				} else if !errors.Is(err, failure) {
					t.Fatalf("unexpected error: %v", err)
				}
			})
		}
	}
	for _, stage := range []string{"begin", "commit"} {
		t.Run(stage, func(t *testing.T) {
			repo, mock := newOrderRepoTest(t)
			begin := mock.ExpectBegin()
			if stage == "begin" {
				begin.WillReturnError(failure)
			} else {
				mock.ExpectQuery(listOrdersQuery).WithArgs(orderTestBuyerID).WillReturnRows(listTestRows(0)).RowsWillBeClosed()
				mock.ExpectCommit().WillReturnError(failure)
			}
			got, err := repo.ListByBuyer(context.Background(), orderTestBuyerID)
			if got != nil || !errors.Is(err, failure) {
				t.Fatalf("got %v, %v", got, err)
			}
		})
	}
}
