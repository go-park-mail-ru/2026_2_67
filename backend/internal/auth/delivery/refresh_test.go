package delivery

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"vibe_market/backend/internal/auth/domain"
)

func allowDeliveryRefresh(t *testing.T, f *deliveryFixture) {
	t.Helper()
	f.tokens.selectToken = func(hash, device string) (domain.RefreshToken, error) {
		if hash != deliveryHash("old-refresh-token") || device != defaultDeviceName {
			t.Errorf("refresh lookup(%q, %q)", hash, device)
		}
		return domain.RefreshToken{UserID: 42, DeviceName: device}, nil
	}
	f.users.selectID = func(id int64) (domain.User, error) {
		if id != 42 {
			t.Errorf("user lookup ID = %d", id)
		}
		return f.users.user, nil
	}
}
func refreshDeliveryRequest() *http.Request {
	req := deliveryRequest("refresh", "")
	req.AddCookie(&http.Cookie{Name: "refreshToken", Value: "old-refresh-token"})
	return req
}

func TestRefreshDeliverySuccess(t *testing.T) {
	f := newDeliveryFixture(t)
	allowDeliveryRefresh(t, f)
	var stored domain.RefreshToken
	revoked := false
	f.tokens.revoke = func(id int64, device string) error {
		if id != 42 || device != defaultDeviceName {
			t.Errorf("Revoke(%d, %q)", id, device)
		}
		revoked = true
		return nil
	}
	inserts := 0
	f.tokens.insert = func(token domain.RefreshToken) error {
		if !revoked {
			t.Error("new token inserted before old token revocation")
		}
		inserts++
		stored = token
		return nil
	}
	w := httptest.NewRecorder()
	f.handler.Refresh(w, refreshDeliveryRequest())
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	response := deliveryJSON(t, w, "accessToken")
	assertDeliveryJWT(t, deliveryJSONString(t, response, "accessToken"), f)
	assertDeliveryCookie(t, w, stored)
	if inserts != 1 {
		t.Errorf("token inserts = %d", inserts)
	}
	if w.Result().Cookies()[0].Value == "old-refresh-token" {
		t.Error("refresh token was not rotated")
	}
}

func TestRefreshDeliveryMissingCookie(t *testing.T) {
	f := newDeliveryFixture(t)
	req := deliveryRequest("refresh", "")
	req.AddCookie(&http.Cookie{Name: "unrelated", Value: "value"})
	w := httptest.NewRecorder()
	f.handler.Refresh(w, req)
	assertDeliveryEmptyError(t, w, http.StatusBadRequest)
}

func TestRefreshDeliveryErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		setup  func(*deliveryFixture)
		status int
	}{
		{"not found", func(f *deliveryFixture) {
			f.tokens.selectToken = func(string, string) (domain.RefreshToken, error) {
				return domain.RefreshToken{}, fmt.Errorf("lookup: %w", domain.ErrRefreshTokenNotFound)
			}
		}, http.StatusForbidden},
		{"expired", func(f *deliveryFixture) {
			f.tokens.selectToken = func(string, string) (domain.RefreshToken, error) {
				return domain.RefreshToken{}, domain.ErrRefreshTokenExpired
			}
		}, http.StatusForbidden},
		{"revoked", func(f *deliveryFixture) {
			f.tokens.selectToken = func(string, string) (domain.RefreshToken, error) {
				return domain.RefreshToken{}, domain.ErrRefreshTokenRevoked
			}
		}, http.StatusForbidden},
		{"lookup failure", func(f *deliveryFixture) {
			f.tokens.selectToken = func(string, string) (domain.RefreshToken, error) {
				return domain.RefreshToken{}, errors.New("database failure")
			}
		}, http.StatusInternalServerError},
		{"user missing", func(f *deliveryFixture) {
			f.users.selectID = func(int64) (domain.User, error) { return domain.User{}, domain.ErrUserIDNotFound }
		}, http.StatusForbidden},
		{"revoke failure", func(f *deliveryFixture) {
			f.tokens.revoke = func(int64, string) error { return errors.New("revoke failed") }
		}, http.StatusInternalServerError},
		{"insert failure", func(f *deliveryFixture) {
			f.tokens.revoke = func(int64, string) error { return nil }
			f.tokens.insert = func(domain.RefreshToken) error { return errors.New("insert failed") }
		}, http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newDeliveryFixture(t)
			allowDeliveryRefresh(t, f)
			tc.setup(f)
			w := httptest.NewRecorder()
			f.handler.Refresh(w, refreshDeliveryRequest())
			assertDeliveryEmptyError(t, w, tc.status)
		})
	}
}
