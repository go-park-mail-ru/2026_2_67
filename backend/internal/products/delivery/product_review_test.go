package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/common"
	orderDomain "vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/products/domain"
	"vibe_market/backend/internal/products/usecase"

	"go.uber.org/zap"
)

type productReviewUsecaseStub struct {
	gotIn     usecase.GetProductReviewsInDTO
	gotSingle usecase.GetProductReviewInDTO
	gotPatch  usecase.PatchProductReviewInDTO
	out       usecase.GetProductReviewsOutDTO
	singleOut usecase.GetProductReviewOutDTO
	patchOut  usecase.PatchProductReviewOutDTO
	err       error
}

func (s *productReviewUsecaseStub) PostProductReview(dto usecase.PostProductReviewInDTO) (usecase.PostProductReviewOutDTO, error) {
	return usecase.PostProductReviewOutDTO{}, nil
}

func (s *productReviewUsecaseStub) GetProductReviews(dto usecase.GetProductReviewsInDTO) (usecase.GetProductReviewsOutDTO, error) {
	s.gotIn = dto
	return s.out, s.err
}

func (s *productReviewUsecaseStub) GetProductReview(dto usecase.GetProductReviewInDTO) (usecase.GetProductReviewOutDTO, error) {
	s.gotSingle = dto
	return s.singleOut, s.err
}

func (s *productReviewUsecaseStub) PatchProductReview(dto usecase.PatchProductReviewInDTO) (usecase.PatchProductReviewOutDTO, error) {
	s.gotPatch = dto
	return s.patchOut, s.err
}

func newTestGetReviewsHandler(stub *productReviewUsecaseStub) http.Handler {
	h := NewProductsHandler(nil, nil, stub, zap.NewNop(), nil)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /products/{id}/reviews", h.GetProductReviews)
	mux.HandleFunc("GET /products/{product_id}/reviews/{buyer_id}", h.GetProductReview)
	mux.HandleFunc("PATCH /products/{product_id}/reviews/{buyer_id}", h.PatchProductReview)
	return common.EndpointLogMiddleware(zap.NewNop())(mux)
}

func TestGetProductReviewsHTTP(t *testing.T) {
	now := time.Date(2026, 11, 24, 15, 0, 0, 0, time.UTC)
	surname := "Иванов"
	reviewText := "Хорошо"

	t.Run("success and cursor roundtrip", func(t *testing.T) {
		stub := &productReviewUsecaseStub{
			out: usecase.GetProductReviewsOutDTO{
				Reviews: []domain.ProductReviewListItem{
					{
						BuyerID:      10,
						BuyerName:    "Иван",
						BuyerSurname: &surname,
						Review:       &reviewText,
						Rating:       5,
						CreatedAt:    now,
					},
					{
						BuyerID:   11,
						BuyerName: "Пётр",
						Rating:    4,
						CreatedAt: now.Add(-time.Minute),
					},
				},
				NextCursor: &domain.ProductReviewCursor{
					CreatedAt: now.Add(-time.Minute),
					BuyerID:   11,
					Rating:    4,
				},
			},
		}
		handler := newTestGetReviewsHandler(stub)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/products/7/reviews?sortBy=desc-rating&limit=2", nil)
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("got status %d, body=%s", w.Code, w.Body.String())
		}

		var resp getProductReviewsResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if len(resp.Reviews) != 2 {
			t.Fatalf("got %d reviews, want 2", len(resp.Reviews))
		}
		if resp.Reviews[0].BuyerName != "Иван" || resp.Reviews[0].BuyerSurname != "Иванов" || resp.Reviews[0].Review != "Хорошо" {
			t.Fatalf("unexpected first item: %+v", resp.Reviews[0])
		}
		if resp.Reviews[1].BuyerPatronymic != "" || resp.Reviews[1].BuyerAvatarURL != "" || resp.Reviews[1].Review != "" {
			t.Fatalf("expected empty strings for nil fields, got %+v", resp.Reviews[1])
		}
		if resp.NextCursor == "" {
			t.Fatal("expected non-empty nextCursor")
		}

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/products/7/reviews?cursor="+url.QueryEscape(resp.NextCursor), nil)
		handler.ServeHTTP(w2, req2)
		if w2.Code != http.StatusOK {
			t.Fatalf("got status %d", w2.Code)
		}
		if stub.gotIn.Cursor == nil || stub.gotIn.Cursor.BuyerID != 11 || stub.gotIn.Cursor.Rating != 4 {
			t.Fatalf("cursor not decoded properly: %+v", stub.gotIn.Cursor)
		}
	})

	t.Run("invalid product id", func(t *testing.T) {
		handler := newTestGetReviewsHandler(&productReviewUsecaseStub{})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/products/abc/reviews", nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want 404", w.Code)
		}
	})

	t.Run("invalid limit", func(t *testing.T) {
		handler := newTestGetReviewsHandler(&productReviewUsecaseStub{})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/products/7/reviews?limit=0", nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want 400", w.Code)
		}
	})

	t.Run("invalid cursor", func(t *testing.T) {
		handler := newTestGetReviewsHandler(&productReviewUsecaseStub{})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/products/7/reviews?cursor=not-valid", nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want 400", w.Code)
		}
	})

	t.Run("invalid sortBy", func(t *testing.T) {
		stub := &productReviewUsecaseStub{err: domain.ErrProductReviewSortByInvalid}
		handler := newTestGetReviewsHandler(stub)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/products/7/reviews?sortBy=bad", nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want 400", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		stub := &productReviewUsecaseStub{err: errors.New("db down")}
		handler := newTestGetReviewsHandler(stub)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/products/7/reviews", nil))
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("got status %d, want 500", w.Code)
		}
	})
}

func withAuthPayload(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.AccessTokenPayloadKey, &authDomain.AccessTokenPayload{UserID: userID})
	return req.WithContext(ctx)
}

func TestGetProductReviewHTTP(t *testing.T) {
	t.Run("success with empty review", func(t *testing.T) {
		stub := &productReviewUsecaseStub{
			singleOut: usecase.GetProductReviewOutDTO{Rating: 5, Review: nil},
		}
		handler := newTestGetReviewsHandler(stub)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/products/7/reviews/12", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("got status %d, want 200", w.Code)
		}
		var resp productReviewResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Rating != 5 || resp.Review != "" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("not found", func(t *testing.T) {
		stub := &productReviewUsecaseStub{err: domain.ErrProductReviewNotFound}
		handler := newTestGetReviewsHandler(stub)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/products/7/reviews/12", nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want 404", w.Code)
		}
	})

	t.Run("invalid buyer id", func(t *testing.T) {
		handler := newTestGetReviewsHandler(&productReviewUsecaseStub{})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/products/7/reviews/abc", nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want 404", w.Code)
		}
	})
}

func TestPatchProductReviewHTTP(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		handler := newTestGetReviewsHandler(&productReviewUsecaseStub{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/products/7/reviews/12", strings.NewReader(`{"rating":5,"review":"ok"}`))
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want 401", w.Code)
		}
	})

	t.Run("forbidden for another buyer", func(t *testing.T) {
		handler := newTestGetReviewsHandler(&productReviewUsecaseStub{})
		w := httptest.NewRecorder()
		req := withAuthPayload(httptest.NewRequest(http.MethodPatch, "/products/7/reviews/12", strings.NewReader(`{"rating":5,"review":"ok"}`)), 99)
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("got status %d, want 403", w.Code)
		}
	})

	t.Run("missing rating returns 400", func(t *testing.T) {
		handler := newTestGetReviewsHandler(&productReviewUsecaseStub{})
		w := httptest.NewRecorder()
		req := withAuthPayload(httptest.NewRequest(http.MethodPatch, "/products/7/reviews/12", strings.NewReader(`{"review":"ok"}`)), 12)
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want 400", w.Code)
		}
	})

	t.Run("422 on invalid rating or order not found", func(t *testing.T) {
		stub := &productReviewUsecaseStub{err: orderDomain.ErrOrdersNotFound}
		handler := newTestGetReviewsHandler(stub)
		w := httptest.NewRecorder()
		req := withAuthPayload(httptest.NewRequest(http.MethodPatch, "/products/7/reviews/12", strings.NewReader(`{"rating":5,"review":"ok"}`)), 12)
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("got status %d, want 422", w.Code)
		}
	})

	t.Run("404 when review not found", func(t *testing.T) {
		stub := &productReviewUsecaseStub{err: domain.ErrProductReviewNotFound}
		handler := newTestGetReviewsHandler(stub)
		w := httptest.NewRecorder()
		req := withAuthPayload(httptest.NewRequest(http.MethodPatch, "/products/7/reviews/12", strings.NewReader(`{"rating":5,"review":""}`)), 12)
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want 404", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		text := "Новый текст"
		stub := &productReviewUsecaseStub{
			patchOut: usecase.PatchProductReviewOutDTO{Rating: 4, Review: &text},
		}
		handler := newTestGetReviewsHandler(stub)
		w := httptest.NewRecorder()
		req := withAuthPayload(httptest.NewRequest(http.MethodPatch, "/products/7/reviews/12", strings.NewReader(`{"rating":4,"review":"Новый текст"}`)), 12)
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("got status %d, want 200", w.Code)
		}
		var resp productReviewResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Rating != 4 || resp.Review != text {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})
}
