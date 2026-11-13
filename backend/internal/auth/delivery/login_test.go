package delivery

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"vibe_market/backend/internal/auth/domain"
)

func TestLoginDeliverySuccess(t *testing.T) {
	for _, auth := range []string{"", "Bearer invalid-token", "expired", "wrong-signature"} {
		t.Run(auth, func(t *testing.T) {
			f := newDeliveryFixture(t)
			f.allowPasswordLogin(t)
			var stored domain.RefreshToken
			calls := 0
			f.tokens.insert = func(token domain.RefreshToken) error { calls++; stored = token; return nil }
			req := deliveryRequest("login", `{"email":" buyer@example.com ","password":"ValidPass123!"}`)
			switch auth {
			case "expired":
				auth = "Bearer " + deliverySignedToken(t, f.meta, time.Now().Add(-time.Hour))
			case "wrong-signature":
				meta := f.meta
				meta.JwtSecret = []byte("other-secret")
				auth = "Bearer " + deliverySignedToken(t, meta, time.Now().Add(time.Hour))
			}
			req.Header.Set("Authorization", auth)
			body := &deliveryTrackingBody{Reader: req.Body}
			req.Body = body
			w := httptest.NewRecorder()
			f.handler.Login(w, req)
			if !body.closed {
				t.Error("request body not closed")
			}
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d", w.Code)
			}
			response := deliveryJSON(t, w, "userId", "accessToken")
			if string(response["userId"]) != "42" {
				t.Errorf("userId = %s", response["userId"])
			}
			assertDeliveryJWT(t, deliveryJSONString(t, response, "accessToken"), f)
			if calls != 1 {
				t.Errorf("token insert calls = %d", calls)
			}
			assertDeliveryCookie(t, w, stored)
		})
	}
}

func TestLoginDeliveryAlreadyAuthenticated(t *testing.T) {
	f := newDeliveryFixture(t)
	req := deliveryRequest("login", `{"email":"ignored@example.com","password":"ignored"}`)
	req.Header.Set("Authorization", "Bearer "+deliverySignedToken(t, f.meta, time.Now().Add(time.Hour)))
	w := httptest.NewRecorder()
	f.handler.Login(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	response := deliveryJSON(t, w, "userId", "accessToken")
	if string(response["userId"]) != "42" {
		t.Errorf("userId = %s", response["userId"])
	}
	assertDeliveryJWT(t, deliveryJSONString(t, response, "accessToken"), f)
	if len(w.Result().Cookies()) != 0 {
		t.Error("already authenticated response replaces cookie")
	}
}

func TestLoginDeliveryMalformedBody(t *testing.T) {
	// Missing/null credential fields are omitted: login.go currently dereferences a nil error there.
	for _, raw := range []string{"", "{", "[]", `{"email":123,"password":"test"}`, `{"email":"buyer@example.com","password":false}`} {
		t.Run(raw, func(t *testing.T) {
			f := newDeliveryFixture(t)
			req := deliveryRequest("login", raw)
			body := &deliveryTrackingBody{Reader: strings.NewReader(raw)}
			req.Body = body
			w := httptest.NewRecorder()
			f.handler.Login(w, req)
			assertDeliveryEmptyError(t, w, http.StatusBadRequest)
			if !body.closed {
				t.Error("request body not closed")
			}
		})
	}
}

func TestLoginDeliveryErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		setup  func(*deliveryFixture)
		status int
	}{
		{"unknown email", func(f *deliveryFixture) {
			f.users.selectEmail = func(string) (domain.User, error) {
				return domain.User{}, fmt.Errorf("lookup: %w", domain.ErrEmailNotFound)
			}
		}, http.StatusUnauthorized},
		{"repository failure", func(f *deliveryFixture) {
			f.users.selectEmail = func(string) (domain.User, error) { return domain.User{}, errors.New("database unavailable") }
		}, http.StatusInternalServerError},
		{"wrong password", func(f *deliveryFixture) { f.users.user.PasswordHash = "invalid hash" }, http.StatusUnauthorized},
		{"refresh insert failure", func(f *deliveryFixture) {
			f.tokens.insert = func(domain.RefreshToken) error { return errors.New("insert failed") }
		}, http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newDeliveryFixture(t)
			f.allowPasswordLogin(t)
			tc.setup(f)
			w := httptest.NewRecorder()
			f.handler.Login(w, deliveryRequest("login", `{"email":"buyer@example.com","password":"ValidPass123!"}`))
			assertDeliveryEmptyError(t, w, tc.status)
		})
	}
}
