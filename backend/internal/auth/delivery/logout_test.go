package delivery

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLogoutDeliveryInvalidAuthorization(t *testing.T) {
	for _, auth := range []string{"", "Basic credentials", "Bearer", "Bearer invalid", "expired", "wrong-signature"} {
		t.Run(auth, func(t *testing.T) {
			f := newDeliveryFixture(t)
			switch auth {
			case "expired":
				auth = "Bearer " + deliverySignedToken(t, f.meta, time.Now().Add(-time.Hour))
			case "wrong-signature":
				meta := f.meta
				meta.JwtSecret = []byte("wrong-secret")
				auth = "Bearer " + deliverySignedToken(t, meta, time.Now().Add(time.Hour))
			}
			req := deliveryRequest("logout", "")
			req.Header.Set("Authorization", auth)
			w := httptest.NewRecorder()
			f.handler.Logout(w, req)
			assertDeliveryEmptyError(t, w, http.StatusForbidden)
		})
	}
}

func TestLogoutDeliveryRevokeFailure(t *testing.T) {
	f := newDeliveryFixture(t)
	calls := 0
	f.tokens.revoke = func(id int64, device string) error {
		calls++
		if id != 42 || device != defaultDeviceName {
			t.Errorf("Revoke(%d, %q)", id, device)
		}
		return errors.New("revoke failed")
	}
	req := deliveryRequest("logout", "")
	req.Header.Set("Authorization", "Bearer "+deliverySignedToken(t, f.meta, time.Now().Add(time.Hour)))
	w := httptest.NewRecorder()
	f.handler.Logout(w, req)
	assertDeliveryEmptyError(t, w, http.StatusForbidden)
	if calls != 1 {
		t.Errorf("revoke calls = %d", calls)
	}
}

func TestLogoutDeliverySuccess(t *testing.T) {
	f := newDeliveryFixture(t)
	calls := 0
	f.tokens.revoke = func(id int64, device string) error {
		calls++
		if id != 42 || device != defaultDeviceName {
			t.Errorf("Revoke(%d, %q)", id, device)
		}
		return nil
	}
	req := deliveryRequest("logout", "")
	req.Header.Set("Authorization", "Bearer "+deliverySignedToken(t, f.meta, time.Now().Add(time.Hour)))
	w := httptest.NewRecorder()
	f.handler.Logout(w, req)
	if w.Code != http.StatusOK || w.Body.Len() != 0 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if calls != 1 {
		t.Errorf("revoke calls = %d", calls)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %v", cookies)
	}
	c := cookies[0]
	if c.Name != "refreshToken" || c.Path != "/api/v1/auth/refresh" || !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
		t.Errorf("unexpected cookie: %+v", c)
	}
	if c.Expires.IsZero() || !c.Expires.Before(time.Now()) {
		t.Errorf("cookie expiry is not in the past: %v", c.Expires)
	}
	// MaxAge is intentionally not asserted: production uses the current second instead of -1.
}
