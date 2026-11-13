package delivery

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"vibe_market/backend/internal/auth/domain"
	buyerDomain "vibe_market/backend/internal/buyers/domain"
)

func allowDeliveryRegister(t *testing.T, f *deliveryFixture) {
	t.Helper()
	ctx := context.Background()
	f.users.begin = func() (context.Context, error) { return ctx, nil }
	f.users.insert = func(got context.Context, email, hash string) (domain.User, error) {
		if got != ctx || email != "buyer@example.com" {
			t.Errorf("unexpected registration arguments: ctx=%v email=%q", got, email)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(deliveryPassword)); err != nil {
			t.Errorf("password hash: %v", err)
		}
		return f.users.user, nil
	}
}

func TestRegisterDeliverySuccess(t *testing.T) {
	f := newDeliveryFixture(t)
	allowDeliveryRegister(t, f)
	var stored domain.RefreshToken
	f.tokens.insert = func(token domain.RefreshToken) error { stored = token; return nil }
	buyerCalls := 0
	f.buyers.insert = func(_ context.Context, buyer buyerDomain.Buyer) error {
		buyerCalls++
		if buyer.UserID != 42 || buyer.Name != "buyer" {
			t.Errorf("buyer = %+v", buyer)
		}
		return nil
	}
	req := deliveryRequest("register", `{"email":" buyer@example.com ","password":"ValidPass123!"}`)
	req.Header.Set("Authorization", "Bearer invalid-token")
	body := &deliveryTrackingBody{Reader: req.Body}
	req.Body = body
	w := httptest.NewRecorder()
	f.handler.Register(w, req)
	if !body.closed {
		t.Error("request body not closed")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; body=%s", w.Code, w.Body.String())
	}
	response := deliveryJSON(t, w, "userId", "accessToken")
	if string(response["userId"]) != "42" {
		t.Errorf("userId = %s", response["userId"])
	}
	assertDeliveryJWT(t, deliveryJSONString(t, response, "accessToken"), f)
	assertDeliveryCookie(t, w, stored)
	if buyerCalls != 1 {
		t.Errorf("buyer inserts = %d", buyerCalls)
	}
}

func TestRegisterDeliveryBadRequest(t *testing.T) {
	for _, raw := range []string{"", "{", "[]", "null", "{}", `{"email":"buyer@example.com"}`, `{"password":"ValidPass123!"}`, `{"email":null,"password":"ValidPass123!"}`, `{"email":"buyer@example.com","password":null}`, `{"email":42,"password":"ValidPass123!"}`, `{"email":"buyer@example.com","password":false}`} {
		t.Run(raw, func(t *testing.T) {
			f := newDeliveryFixture(t)
			req := deliveryRequest("register", raw)
			body := &deliveryTrackingBody{Reader: req.Body}
			req.Body = body
			w := httptest.NewRecorder()
			f.handler.Register(w, req)
			assertDeliveryEmptyError(t, w, http.StatusBadRequest)
			if !body.closed {
				t.Error("request body not closed")
			}
		})
	}
}

func TestRegisterDeliveryAlreadyAuthenticated(t *testing.T) {
	f := newDeliveryFixture(t)
	req := deliveryRequest("register", `{"email":"buyer@example.com","password":"ValidPass123!"}`)
	req.Header.Set("Authorization", "Bearer "+deliverySignedToken(t, f.meta, time.Now().Add(time.Hour)))
	w := httptest.NewRecorder()
	f.handler.Register(w, req)
	assertDeliveryEmptyError(t, w, http.StatusBadRequest)
}

func TestRegisterDeliveryValidation(t *testing.T) {
	for _, tc := range []struct {
		email, password           string
		emailError, passwordError bool
	}{
		{"invalid", deliveryPassword, true, false},
		{"buyer@example.com", "weak", false, true},
		{"", "", true, true},
	} {
		t.Run(tc.email+"/"+tc.password, func(t *testing.T) {
			f := newDeliveryFixture(t)
			w := httptest.NewRecorder()
			f.handler.Register(w, deliveryRequest("register", fmt.Sprintf(`{"email":%q,"password":%q}`, tc.email, tc.password)))
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d", w.Code)
			}
			response := deliveryJSON(t, w, "emailErrMessage", "passwordErrMessage")
			if (deliveryJSONString(t, response, "emailErrMessage") != "") != tc.emailError {
				t.Error("unexpected email validation message")
			}
			if (deliveryJSONString(t, response, "passwordErrMessage") != "") != tc.passwordError {
				t.Error("unexpected password validation message")
			}
			if len(w.Result().Cookies()) != 0 {
				t.Error("validation response sets cookie")
			}
		})
	}
}

func TestRegisterDeliveryConflict(t *testing.T) {
	f := newDeliveryFixture(t)
	allowDeliveryRegister(t, f)
	f.users.insert = func(context.Context, string, string) (domain.User, error) {
		return domain.User{}, fmt.Errorf("insert: %w", domain.ErrEmailAlreadyExists)
	}
	w := httptest.NewRecorder()
	f.handler.Register(w, deliveryRequest("register", `{"email":"buyer@example.com","password":"ValidPass123!"}`))
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d", w.Code)
	}
	response := deliveryJSON(t, w, "emailErrMessage", "passwordErrMessage")
	if deliveryJSONString(t, response, "emailErrMessage") == "" || deliveryJSONString(t, response, "passwordErrMessage") != "" {
		t.Errorf("unexpected conflict response: %v", response)
	}
	if len(w.Result().Cookies()) != 0 {
		t.Error("conflict response sets cookies")
	}
}

func TestRegisterDeliveryInternalErrors(t *testing.T) {
	for _, stage := range []string{"begin", "refresh insert"} {
		t.Run(stage, func(t *testing.T) {
			f := newDeliveryFixture(t)
			allowDeliveryRegister(t, f)
			if stage == "begin" {
				f.users.begin = func() (context.Context, error) { return nil, errors.New("begin failed") }
			} else {
				f.tokens.insert = func(domain.RefreshToken) error { return errors.New("insert failed") }
			}
			w := httptest.NewRecorder()
			f.handler.Register(w, deliveryRequest("register", `{"email":"buyer@example.com","password":"ValidPass123!"}`))
			assertDeliveryEmptyError(t, w, http.StatusInternalServerError)
		})
	}
}
