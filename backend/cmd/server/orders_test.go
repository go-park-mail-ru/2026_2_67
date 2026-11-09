package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrdersRoutes(t *testing.T) {
	for _, path := range []string{"/api/v1/orders", "/api/v1/orders/"} {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			t.Run(method+path, func(t *testing.T) {
				mainMux := http.NewServeMux()
				ordersMux := http.NewServeMux()
				called := false
				req := httptest.NewRequest(method, path+"?buyerId=99", nil)
				ordersMux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
					called = true
					if r.Method != method || r.URL.Query().Get("buyerId") != "99" || r.Context() != req.Context() {
						t.Error("request method, query or context changed")
					}
					w.WriteHeader(http.StatusNoContent)
				})
				registerOrdersRoutes(mainMux, ordersMux)
				w := httptest.NewRecorder()
				mainMux.ServeHTTP(w, req)
				if w.Code != http.StatusNoContent || !called || w.Header().Get("Location") != "" {
					t.Fatalf("status=%d called=%v location=%q", w.Code, called, w.Header().Get("Location"))
				}
				if req.URL.Path != path {
					t.Error("original request path was modified")
				}
			})
		}
	}
	mainMux := http.NewServeMux()
	ordersMux := http.NewServeMux()
	ordersMux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) { t.Fatal("unexpected handler call") })
	registerOrdersRoutes(mainMux, ordersMux)
	w := httptest.NewRecorder()
	mainMux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/orders/unknown", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status=%d", w.Code)
	}
}
