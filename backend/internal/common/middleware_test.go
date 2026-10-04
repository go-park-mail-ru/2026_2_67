package common

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireHTTPMethod(t *testing.T) {
	called := false
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	chain := RequireHTTPMethod(http.MethodGet)(dummyHandler)

	// Test correct method
	reqGet := httptest.NewRequest(http.MethodGet, "/test", nil)
	rrGet := httptest.NewRecorder()
	chain.ServeHTTP(rrGet, reqGet)

	if rrGet.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rrGet.Code)
	}
	if !called {
		t.Errorf("expected next handler to be called")
	}

	// Test incorrect method
	called = false
	reqPost := httptest.NewRequest(http.MethodPost, "/test", nil)
	rrPost := httptest.NewRecorder()
	chain.ServeHTTP(rrPost, reqPost)

	if rrPost.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rrPost.Code)
	}
	if called {
		t.Errorf("expected next handler not to be called")
	}
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	called := false
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	cors := CORSMiddleware(dummyHandler)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/products", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
	rr := httptest.NewRecorder()

	cors.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d for OPTIONS, got %d", http.StatusNoContent, rr.Code)
	}

	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:3000" {
		t.Errorf("expected Allow-Origin 'http://localhost:3000', got %s", origin)
	}

	if creds := rr.Header().Get("Access-Control-Allow-Credentials"); creds != "true" {
		t.Errorf("expected Allow-Credentials 'true', got %s", creds)
	}

	if headers := rr.Header().Get("Access-Control-Allow-Headers"); headers != "Content-Type, Authorization" {
		t.Errorf("expected Allow-Headers to match request headers, got %s", headers)
	}

	if called {
		t.Errorf("expected dummyHandler not to be called on OPTIONS preflight")
	}
}

func TestCORSMiddleware_ActualRequestWithOrigin(t *testing.T) {
	called := false
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	cors := CORSMiddleware(dummyHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rr := httptest.NewRecorder()

	cors.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:5173" {
		t.Errorf("expected Allow-Origin 'http://localhost:5173', got %s", origin)
	}

	if creds := rr.Header().Get("Access-Control-Allow-Credentials"); creds != "true" {
		t.Errorf("expected Allow-Credentials 'true', got %s", creds)
	}

	if !called {
		t.Errorf("expected dummyHandler to be called")
	}
}

func TestCORSMiddleware_ActualRequestWithoutOrigin(t *testing.T) {
	called := false
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	cors := CORSMiddleware(dummyHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rr := httptest.NewRecorder()

	cors.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("expected Allow-Origin '*', got %s", origin)
	}

	if !called {
		t.Errorf("expected dummyHandler to be called")
	}
}
