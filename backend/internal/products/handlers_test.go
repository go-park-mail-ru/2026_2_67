package products

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockStorage struct {
	products []Product
	err      error
}

func (m *mockStorage) SelectAllProducts() ([]Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.products, nil
}

func (m *mockStorage) InsertProduct(product Product) (Product, error) {
	if m.err != nil {
		return Product{}, m.err
	}
	m.products = append(m.products, product)
	return product, nil
}

func TestGetProductsHandler_Success(t *testing.T) {
	mock := &mockStorage{
		products: []Product{
			{
				ID:           1,
				Name:         "Телевизор",
				PictureURLs:  []string{"https://example.com/tv.jpg"},
				Price:        45000,
				Rating:       4.6,
				ReviewsCount: 35,
			},
		},
	}

	handler := MakeProductsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rr := httptest.NewRecorder()

	handler.GetProductsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var response []ProductResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Fatalf("expected 1 product, got %d", len(response))
	}

	if response[0].ProductName != "Телевизор" {
		t.Errorf("expected productName 'Телевизор', got %s", response[0].ProductName)
	}
	if len(response[0].ProductPictureURLs) != 1 || response[0].ProductPictureURLs[0] != "https://example.com/tv.jpg" {
		t.Errorf("unexpected picture urls: %+v", response[0].ProductPictureURLs)
	}
	if response[0].ProductPrice != 45000 {
		t.Errorf("expected productPrice 45000, got %d", response[0].ProductPrice)
	}
	if response[0].ProductRating != 4.6 {
		t.Errorf("expected productRating 4.6, got %f", response[0].ProductRating)
	}
	if response[0].ProductReviewsCount != 35 {
		t.Errorf("expected productReviewsCount 35, got %d", response[0].ProductReviewsCount)
	}
}

func TestGetProductsHandler_Empty(t *testing.T) {
	mock := &mockStorage{
		products: []Product{},
	}

	handler := MakeProductsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rr := httptest.NewRecorder()

	handler.GetProductsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	expectedBody := "[]\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rr.Body.String())
	}
}

func TestGetProductsHandler_Error(t *testing.T) {
	mock := &mockStorage{
		err: errors.New("db error"),
	}

	handler := MakeProductsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rr := httptest.NewRecorder()

	handler.GetProductsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status code %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}
