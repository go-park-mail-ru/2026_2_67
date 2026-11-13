package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"vibe_market/backend/internal/categories/domain"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

type categoryRepoStub struct {
	calls      []domain.CategoriesSearchParams
	categories []domain.Category
	err        error
}

func (s *categoryRepoStub) Search(params domain.CategoriesSearchParams) ([]domain.Category, error) {
	s.calls = append(s.calls, params)
	return s.categories, s.err
}

func requestCategories(t *testing.T, repo *categoryRepoStub, query string, wantParams domain.CategoriesSearchParams) *httptest.ResponseRecorder {
	t.Helper()

	logger := zap.NewNop()
	h := NewCategoriesHandler(repo, logger)
	r := httptest.NewRequest(http.MethodGet, "/"+query, nil)
	r = r.WithContext(context.WithValue(r.Context(), common.LoggerContextKey, logger))
	w := httptest.NewRecorder()
	h.Categories(w, r)

	if len(repo.calls) != 1 || repo.calls[0] != wantParams {
		t.Errorf("Search calls = %+v, want one call with %+v", repo.calls, wantParams)
	}
	return w
}

func assertCategoriesJSON(t *testing.T, w *httptest.ResponseRecorder, want string) {
	t.Helper()

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", w.Code, http.StatusOK, w.Body.String())
	}
	if got := w.Result().Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	var gotJSON, wantJSON any
	if err := json.Unmarshal(w.Body.Bytes(), &gotJSON); err != nil {
		t.Fatalf("decode response JSON: %v; body = %q", err, w.Body.String())
	}
	if err := json.Unmarshal([]byte(want), &wantJSON); err != nil {
		t.Fatalf("decode expected JSON: %v", err)
	}
	if !reflect.DeepEqual(gotJSON, wantJSON) {
		t.Errorf("body = %s, want %s", w.Body.String(), want)
	}
}

func TestCategoriesQuery(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query string
		want  domain.CategoriesSearchParams
	}{
		{"popular with limit", "?sortBy=popular&limit=12", domain.CategoriesSearchParams{SortBy: domain.SortByPopular, Limit: 12}},
		{"new without limit", "?sortBy=new", domain.CategoriesSearchParams{SortBy: domain.SortByNew}},
		{"limit without sort", "?limit=7", domain.CategoriesSearchParams{Limit: 7}},
		{"no parameters", "", domain.CategoriesSearchParams{}},
		{"empty parameters", "?sortBy=&limit=", domain.CategoriesSearchParams{}},
		{"zero limit", "?limit=0", domain.CategoriesSearchParams{}},
		{"non-numeric limit", "?sortBy=new&limit=abc", domain.CategoriesSearchParams{SortBy: domain.SortByNew}},
		{"negative limit", "?limit=-1", domain.CategoriesSearchParams{}},
		{"maximum limit", "?limit=4294967295", domain.CategoriesSearchParams{Limit: 4294967295}},
		{"overflowing limit", "?limit=4294967296", domain.CategoriesSearchParams{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := requestCategories(t, &categoryRepoStub{}, tc.query, tc.want)
			assertCategoriesJSON(t, w, `{"categories":[]}`)
		})
	}
}

func TestCategoriesJSONPreservesRepositoryOrder(t *testing.T) {
	picture := "/pictures/books.jpg"
	repo := &categoryRepoStub{categories: []domain.Category{
		{ID: 8, Category: "Книги", PictureURL: &picture},
		{ID: 2, Category: "Игры", PictureURL: nil},
		{ID: 5, Category: "Комиксы", PictureURL: &picture},
	}}
	w := requestCategories(t, repo, "?sortBy=popular", domain.CategoriesSearchParams{SortBy: domain.SortByPopular})

	// A nil repository picture is rendered as an empty string by the current API.
	assertCategoriesJSON(t, w, `{"categories":[
		{"id":8,"category":"Книги","pictureUrl":"/pictures/books.jpg"},
		{"id":2,"category":"Игры","pictureUrl":""},
		{"id":5,"category":"Комиксы","pictureUrl":"/pictures/books.jpg"}
	]}`)
}

func TestCategoriesEmptyRepositoryResult(t *testing.T) {
	repo := &categoryRepoStub{categories: []domain.Category{}}
	w := requestCategories(t, repo, "", domain.CategoriesSearchParams{})
	assertCategoriesJSON(t, w, `{"categories":[]}`)
}

func TestCategoriesInvalidSortBy(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"sentinel", domain.ErrInvalidSortBy},
		{"wrapped sentinel", fmt.Errorf("search categories: %w", domain.ErrInvalidSortBy)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &categoryRepoStub{
				categories: []domain.Category{{ID: 8, Category: "Книги"}},
				err:        tc.err,
			}
			w := requestCategories(t, repo, "?sortBy=unknown&limit=3", domain.CategoriesSearchParams{SortBy: "unknown", Limit: 3})
			assertCategoriesJSON(t, w, `{"categories":[]}`)
		})
	}
}

func TestCategoriesRepositoryError(t *testing.T) {
	repo := &categoryRepoStub{
		categories: []domain.Category{{ID: 8, Category: "Книги"}},
		err:        errors.New("database unavailable"),
	}
	w := requestCategories(t, repo, "", domain.CategoriesSearchParams{})
	if w.Code != http.StatusInternalServerError || w.Body.Len() != 0 {
		t.Errorf("status = %d, body = %q; want %d with an empty body", w.Code, w.Body.String(), http.StatusInternalServerError)
	}
}
