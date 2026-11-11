package delivery

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"vibe_market/backend/internal/orders/domain"

	"go.uber.org/zap"
)

func listOrdersTestRequest(t *testing.T, repo *orderRepoStub, authorization string) (*httptest.ResponseRecorder, context.Context) {
	t.Helper()
	h := NewOrdersHandler(repo, orderTestSecret, zap.NewNop())
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "list-orders-context")
	r := httptest.NewRequest(http.MethodGet, "/api/v1/orders?buyerId=99", nil).WithContext(ctx)
	if authorization != "" {
		r.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	h.ListOrders(w, r)
	return w, ctx
}

func TestListOrdersSuccess(t *testing.T) {
	created := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	updated := created.Add(time.Hour)
	promoStart := created.Add(-24 * time.Hour)
	promoEnd := created.Add(24 * time.Hour)
	categoryID := int64(7)
	description := "Описание товара"
	emptyDescription := ""
	repo := &orderRepoStub{orders: []domain.OrderDetails{
		{
			ID: 101, Status: domain.OrderStatus{ID: 2, Name: "В пути"},
			CreatedAt: created, UpdatedAt: updated,
			PickupPoint: domain.OrderPickupPoint{ID: 17, Longitude: 37.625, Latitude: 55.75, StartTime: "09:00", EndTime: "21:00"},
			Promocode:   &domain.OrderPromocode{ID: 9, Code: "SALE", Discount: 25, AvailableCategoryID: &categoryID, StartDatetime: promoStart, EndDatetime: promoEnd},
			Products: []domain.OrderProductDetails{
				{
					ID: 201, ProductID: 301, Name: "Товар", Description: &description,
					Category: domain.OrderCategory{ID: 7, Name: "Категория"}, Seller: domain.OrderSeller{ID: 8, Name: "Продавец"},
					PictureURLs: []string{"/pictures/first.jpg", "/pictures/second.jpg"}, Count: 3, Price: 4000000000,
				},
				{
					ID: 202, ProductID: 302, Name: "Другой товар", Description: nil,
					Category: domain.OrderCategory{ID: 10, Name: "Другая категория"}, Seller: domain.OrderSeller{ID: 11, Name: "Магазин"},
					PictureURLs: nil, Count: 2, Price: 125,
				},
			},
			Notifications: []domain.OrderNotification{
				{ID: 401, Status: domain.OrderStatus{ID: 1, Name: "Создан"}, IsRead: true, CreatedAt: created},
				{ID: 402, Status: domain.OrderStatus{ID: 2, Name: "В пути"}, IsRead: false, CreatedAt: updated},
			},
		},
		{
			ID: 102, Status: domain.OrderStatus{ID: 1, Name: "Создан"}, CreatedAt: created, UpdatedAt: created,
			PickupPoint: domain.OrderPickupPoint{ID: 18, Longitude: 30.25, Latitude: 60.5, StartTime: "10:00", EndTime: "20:00"},
			Promocode:   &domain.OrderPromocode{ID: 12, Code: "ALL", Discount: 10, AvailableCategoryID: nil, StartDatetime: promoStart, EndDatetime: promoEnd},
			Products: []domain.OrderProductDetails{
				{ID: 203, ProductID: 303, Name: "Без описания", Description: &emptyDescription,
					Category: domain.OrderCategory{ID: 7, Name: "Категория"}, Seller: domain.OrderSeller{ID: 8, Name: "Продавец"},
					PictureURLs: []string{}, Count: 1, Price: 999},
			},
			Notifications: []domain.OrderNotification{},
		},
		{
			ID: 103, Status: domain.OrderStatus{ID: 3, Name: "Отменён"}, CreatedAt: created, UpdatedAt: updated,
			PickupPoint: domain.OrderPickupPoint{ID: 17, Longitude: 37.625, Latitude: 55.75, StartTime: "09:00", EndTime: "21:00"},
			Promocode:   nil, Products: nil, Notifications: nil,
		},
	}}
	auth := orderTestToken(t, orderTestSecret, time.Now().Add(time.Hour))
	w, ctx := listOrdersTestRequest(t, repo, auth)
	// Prices are order-time snapshots: neither promo discount nor current catalog prices apply.
	assertOrderJSON(t, w, http.StatusOK, `{"orders":[
		{"id":101,"status":{"id":2,"name":"В пути"},"createdAt":"2026-01-02T03:04:05Z","updatedAt":"2026-01-02T04:04:05Z",
		"pickupPoint":{"id":17,"longitude":37.625,"latitude":55.75,"startTime":"09:00","endTime":"21:00"},
		"promocode":{"id":9,"code":"SALE","discount":25,"availableCategoryId":7,"startDatetime":"2026-01-01T03:04:05Z","endDatetime":"2026-01-03T03:04:05Z"},
		"products":[
			{"id":201,"productId":301,"name":"Товар","description":"Описание товара","category":{"id":7,"name":"Категория"},"seller":{"id":8,"name":"Продавец"},"pictureUrls":["/pictures/first.jpg","/pictures/second.jpg"],"count":3,"unitPrice":4000000000,"totalPrice":12000000000},
			{"id":202,"productId":302,"name":"Другой товар","description":null,"category":{"id":10,"name":"Другая категория"},"seller":{"id":11,"name":"Магазин"},"pictureUrls":[],"count":2,"unitPrice":125,"totalPrice":250}],
		"notifications":[{"id":401,"status":{"id":1,"name":"Создан"},"isRead":true,"createdAt":"2026-01-02T03:04:05Z"},{"id":402,"status":{"id":2,"name":"В пути"},"isRead":false,"createdAt":"2026-01-02T04:04:05Z"}],"totalCount":5,"totalPrice":12000000250},
		{"id":102,"status":{"id":1,"name":"Создан"},"createdAt":"2026-01-02T03:04:05Z","updatedAt":"2026-01-02T03:04:05Z",
		"pickupPoint":{"id":18,"longitude":30.25,"latitude":60.5,"startTime":"10:00","endTime":"20:00"},
		"promocode":{"id":12,"code":"ALL","discount":10,"availableCategoryId":null,"startDatetime":"2026-01-01T03:04:05Z","endDatetime":"2026-01-03T03:04:05Z"},
		"products":[{"id":203,"productId":303,"name":"Без описания","description":"","category":{"id":7,"name":"Категория"},"seller":{"id":8,"name":"Продавец"},"pictureUrls":[],"count":1,"unitPrice":999,"totalPrice":999}],"notifications":[],"totalCount":1,"totalPrice":999},
		{"id":103,"status":{"id":3,"name":"Отменён"},"createdAt":"2026-01-02T03:04:05Z","updatedAt":"2026-01-02T04:04:05Z",
		"pickupPoint":{"id":17,"longitude":37.625,"latitude":55.75,"startTime":"09:00","endTime":"21:00"},"promocode":null,"products":[],"notifications":[],"totalCount":0,"totalPrice":0}
	]}`)
	if repo.calls != 1 || repo.buyerID != 42 || repo.ctx != ctx {
		t.Errorf("unexpected repository call: %+v", repo)
	}
}

func TestListOrdersEmpty(t *testing.T) {
	auth := orderTestToken(t, orderTestSecret, time.Now().Add(time.Hour))
	for _, tc := range []struct {
		name   string
		orders []domain.OrderDetails
	}{
		{"nil slice", nil},
		{"empty slice", []domain.OrderDetails{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &orderRepoStub{orders: tc.orders}
			w, ctx := listOrdersTestRequest(t, repo, auth)
			assertOrderJSON(t, w, http.StatusOK, `{"orders":[]}`)
			if repo.calls != 1 || repo.buyerID != 42 || repo.ctx != ctx {
				t.Errorf("unexpected repository call: %+v", repo)
			}
		})
	}
}

func TestListOrdersAuthentication(t *testing.T) {
	for _, tc := range []struct {
		name, auth string
		status     int
	}{
		{"missing header", "", http.StatusBadRequest},
		{"wrong scheme", "Basic token", http.StatusForbidden},
		{"malformed token", "Bearer invalid", http.StatusForbidden},
		{"wrong signature", orderTestToken(t, []byte("wrong-secret"), time.Now().Add(time.Hour)), http.StatusForbidden},
		{"expired token", orderTestToken(t, orderTestSecret, time.Now().Add(-time.Hour)), http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &orderRepoStub{}
			w, _ := listOrdersTestRequest(t, repo, tc.auth)
			if w.Code != tc.status || w.Body.Len() != 0 || repo.calls != 0 {
				t.Fatalf("status = %d, want %d; body = %q, repository calls = %d", w.Code, tc.status, w.Body.String(), repo.calls)
			}
		})
	}
}

func TestListOrdersInternalError(t *testing.T) {
	repo := &orderRepoStub{orders: []domain.OrderDetails{{ID: 101}}, err: errors.New("database unavailable")}
	auth := orderTestToken(t, orderTestSecret, time.Now().Add(time.Hour))
	w, ctx := listOrdersTestRequest(t, repo, auth)
	if w.Code != http.StatusInternalServerError || w.Body.Len() != 0 {
		t.Fatalf("status = %d, body = %q", w.Code, w.Body.String())
	}
	if repo.calls != 1 || repo.buyerID != 42 || repo.ctx != ctx {
		t.Errorf("unexpected repository call: %+v", repo)
	}
}
