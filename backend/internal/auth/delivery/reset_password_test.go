package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"vibe_market/backend/internal/auth/domain"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type dummyUserRepo struct {
	updatedID   int64
	updatedHash string
	updateErr   error
}

func (d *dummyUserRepo) BeginTx() (context.Context, error)                             { return nil, nil }
func (d *dummyUserRepo) InsertTx(ctx context.Context, email, pw string) (domain.User, error) {
	return domain.User{}, nil
}
func (d *dummyUserRepo) RollbackTx(ctx context.Context)                                {}
func (d *dummyUserRepo) CommitTx(ctx context.Context)                                  {}
func (d *dummyUserRepo) UpdateEmailByUserIDTx(ctx context.Context, email string, id int64) error {
	return nil
}
func (d *dummyUserRepo) SelectByEmail(email string) (domain.User, error) {
	return domain.User{}, nil
}
func (d *dummyUserRepo) SelectByUserID(id int64) (domain.User, error) {
	return domain.User{}, nil
}
func (d *dummyUserRepo) UpdatePasswordByUserID(id int64, passwordHash string) error {
	d.updatedID = id
	d.updatedHash = passwordHash
	return d.updateErr
}

func makeTestToken(userID int64, secret []byte) string {
	claims := domain.AccessTokenPayload{
		User: domain.User{
			UserID: userID,
			Email:  "test@example.com",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := t.SignedString(secret)
	return tokenStr
}

func TestResetPasswordDelivery(t *testing.T) {
	jwtSecret := []byte("secret_test_key_12345")
	meta := domain.JwtAuthMeta{
		JwtIssuer: "VibeMarket",
		JwtSecret: jwtSecret,
	}

	repo := &dummyUserRepo{}
	logger := zap.NewNop()
	handler := NewAuthHandler(meta, repo, nil, nil, logger)

	validToken := makeTestToken(10, jwtSecret)

	t.Run("Missing Authorization Header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password", bytes.NewBufferString(`{"password":"ValidPass123!"}`))
		w := httptest.NewRecorder()

		handler.ResetPassword(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password", bytes.NewBufferString(`{"password":"ValidPass123!"}`))
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		handler.ResetPassword(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("Invalid Body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password", bytes.NewBufferString(`{invalid json`))
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()

		handler.ResetPassword(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Weak Password Returns 422", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password", bytes.NewBufferString(`{"password":"weak"}`))
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()

		handler.ResetPassword(w, req)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}

		var resp resetPasswordErrResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.PasswordErrMsg == "" {
			t.Fatal("expected non-empty PasswordErrMsg")
		}
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password", bytes.NewBufferString(`{"password":"ValidPass123!"}`))
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()

		handler.ResetPassword(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if repo.updatedID != 10 {
			t.Fatalf("expected user id 10, got %d", repo.updatedID)
		}
		if repo.updatedHash == "" {
			t.Fatal("expected updated password hash")
		}
	})
}
