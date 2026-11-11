package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"vibe_market/backend/internal/auth/domain"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func TestParseAccessTokenPayload(t *testing.T) {
	now := time.Now()
	validClaims := domain.AccessTokenPayload{
		User: domain.User{
			UserID: 42,
			Email:  "user@example.com",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			Subject:   "42",
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	tests := []struct {
		name          string
		authorization string
		secret        []byte
		wantErr       bool
		wantClaims    *domain.AccessTokenPayload
	}{
		{
			name:    "missing authorization",
			secret:  []byte(testSecret),
			wantErr: true,
		},
		{
			name:          "malformed token",
			authorization: "Bearer not-a-jwt",
			secret:        []byte(testSecret),
			wantErr:       true,
		},
		{
			name:          "basic authorization scheme",
			authorization: "Basic dXNlcjpwYXNz",
			secret:        []byte(testSecret),
			wantErr:       true,
		},
		{
			name:          "basic token",
			authorization: "Bearer " + signedToken(t, domain.AccessTokenPayload{}, []byte(testSecret)),
			secret:        []byte(testSecret),
			wantClaims:    &domain.AccessTokenPayload{},
		},
		{
			name:          "wrong secret",
			authorization: "Bearer " + signedToken(t, validClaims, []byte("another-secret")),
			secret:        []byte(testSecret),
			wantErr:       true,
		},
		{
			name: "expired token",
			authorization: "Bearer " + signedToken(t, domain.AccessTokenPayload{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(now.Add(-time.Minute)),
				},
			}, []byte(testSecret)),
			secret:  []byte(testSecret),
			wantErr: true,
		},
		{
			name:          "invalid algorithm",
			authorization: "Bearer " + noneToken(t),
			secret:        []byte(testSecret),
			wantErr:       true,
		},
		{
			name:          "invalid bearer spacing",
			authorization: "Bearer  " + signedToken(t, validClaims, []byte(testSecret)),
			secret:        []byte(testSecret),
			wantErr:       true,
		},
		{
			name:          "tab instead of bearer space",
			authorization: "Bearer\t" + signedToken(t, validClaims, []byte(testSecret)),
			secret:        []byte(testSecret),
			wantErr:       true,
		},
		{
			name:          "empty bearer token",
			authorization: "Bearer ",
			secret:        []byte(testSecret),
			wantErr:       true,
		},
		{
			name:          "case insensitive bearer",
			authorization: "bEaReR " + signedToken(t, validClaims, []byte(testSecret)),
			secret:        []byte(testSecret),
			wantClaims:    &validClaims,
		},
		{
			name:          "valid claims",
			authorization: "Bearer " + signedToken(t, validClaims, []byte(testSecret)),
			secret:        []byte(testSecret),
			wantClaims:    &validClaims,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authorization != "" {
				r.Header.Set("Authorization", tt.authorization)
			}

			got, err := ParseAccessTokenPayload(r, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseAccessTokenPayload() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if got != nil {
					t.Fatalf("ParseAccessTokenPayload() payload = %#v, want nil", got)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.wantClaims) {
				t.Fatalf("ParseAccessTokenPayload() payload = %#v, want %#v", got, tt.wantClaims)
			}
		})
	}
}

func TestAccessTokenPayloadMiddleware(t *testing.T) {
	validPayload := &domain.AccessTokenPayload{User: domain.User{UserID: 7, Email: "valid@example.com"}}
	validToken := signedToken(t, *validPayload, []byte(testSecret))

	tests := []struct {
		name          string
		authorization string
		wantPayload   *domain.AccessTokenPayload
	}{
		{
			name:          "valid payload in context",
			authorization: "Bearer " + validToken,
			wantPayload:   validPayload,
		},
		{
			name:          "invalid context typed nil",
			authorization: "Bearer malformed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			handler := AccessTokenPayloadMiddleware([]byte(testSecret))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				value, ok := r.Context().Value(AccessTokenPayloadKey).(*domain.AccessTokenPayload)
				if !ok {
					t.Errorf("context payload type assertion failed")
				}
				if !reflect.DeepEqual(value, tt.wantPayload) {
					t.Errorf("context payload = %#v, want %#v", value, tt.wantPayload)
				}
				w.WriteHeader(http.StatusNoContent)
			}))

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("Authorization", tt.authorization)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, r)

			if calls.Load() != 1 {
				t.Fatalf("next called %d times, want 1", calls.Load())
			}
			if recorder.Code != http.StatusNoContent {
				t.Fatalf("response status = %d, want %d", recorder.Code, http.StatusNoContent)
			}
		})
	}
}

func signedToken(t *testing.T, claims domain.AccessTokenPayload, secret []byte) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func noneToken(t *testing.T) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, domain.AccessTokenPayload{}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none token: %v", err)
	}
	return token
}

func TestParseAccessTokenPayloadRejectsRSA(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, domain.AccessTokenPayload{}).SignedString(key)
	if err != nil {
		t.Fatalf("sign RSA token: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	if _, err := ParseAccessTokenPayload(r, []byte(testSecret)); err == nil {
		t.Fatal("ParseAccessTokenPayload() error = nil, want invalid algorithm error")
	}
}
