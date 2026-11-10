package delivery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"vibe_market/backend/internal/auth/domain"
	buyerDomain "vibe_market/backend/internal/buyers/domain"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const deliveryPassword = "ValidPass123!"

// Embedded interfaces leave unrelated repository operations out of these delivery stubs.
// Every operation reached by the four handlers is implemented below.
type deliveryUserStub struct {
	domain.UserRepo
	t           *testing.T
	user        domain.User
	selectEmail func(string) (domain.User, error)
	selectID    func(int64) (domain.User, error)
	begin       func() (context.Context, error)
	insert      func(context.Context, string, string) (domain.User, error)
}

func (s *deliveryUserStub) SelectByEmail(email string) (domain.User, error) {
	if s.selectEmail == nil {
		s.t.Fatalf("unexpected SelectByEmail(%q)", email)
	}
	return s.selectEmail(email)
}
func (s *deliveryUserStub) SelectByUserID(id int64) (domain.User, error) {
	if s.selectID == nil {
		s.t.Fatalf("unexpected SelectByUserID(%d)", id)
	}
	return s.selectID(id)
}
func (s *deliveryUserStub) BeginTx() (context.Context, error) {
	if s.begin == nil {
		s.t.Fatal("unexpected BeginTx")
	}
	return s.begin()
}
func (s *deliveryUserStub) InsertTx(ctx context.Context, email, hash string) (domain.User, error) {
	if s.insert == nil {
		s.t.Fatal("unexpected InsertTx")
	}
	return s.insert(ctx, email, hash)
}
func (s *deliveryUserStub) RollbackTx(context.Context) {}
func (s *deliveryUserStub) CommitTx(context.Context)   {}

type deliveryTokenStub struct {
	t           *testing.T
	insert      func(domain.RefreshToken) error
	revoke      func(int64, string) error
	selectToken func(string, string) (domain.RefreshToken, error)
}

func (s *deliveryTokenStub) Insert(token domain.RefreshToken) error {
	if s.insert == nil {
		s.t.Fatal("unexpected token Insert")
	}
	return s.insert(token)
}
func (s *deliveryTokenStub) Revoke(id int64, device string) error {
	if s.revoke == nil {
		s.t.Fatal("unexpected Revoke")
	}
	return s.revoke(id, device)
}
func (s *deliveryTokenStub) SelectByHashAndDeviceName(hash, device string) (domain.RefreshToken, error) {
	if s.selectToken == nil {
		s.t.Fatal("unexpected SelectByHashAndDeviceName")
	}
	return s.selectToken(hash, device)
}

type deliveryBuyerStub struct {
	buyerDomain.BuyerRepo
	t      *testing.T
	insert func(context.Context, buyerDomain.Buyer) error
}

func (s *deliveryBuyerStub) InsertTx(ctx context.Context, buyer buyerDomain.Buyer) error {
	if s.insert == nil {
		s.t.Fatal("unexpected buyer InsertTx")
	}
	return s.insert(ctx, buyer)
}

type deliveryFixture struct {
	handler *AuthHandler
	users   *deliveryUserStub
	tokens  *deliveryTokenStub
	buyers  *deliveryBuyerStub
	meta    domain.JwtAuthMeta
}

func newDeliveryFixture(t *testing.T) *deliveryFixture {
	t.Helper()
	meta := domain.JwtAuthMeta{JwtIssuer: "delivery-tests", JwtSecret: []byte("delivery-test-signing-secret")}
	users := &deliveryUserStub{t: t, user: domain.User{UserID: 42, Email: "buyer@example.com"}}
	tokens := &deliveryTokenStub{t: t}
	buyers := &deliveryBuyerStub{t: t}
	return &deliveryFixture{NewAuthHandler(meta, users, tokens, buyers, zap.NewNop()), users, tokens, buyers, meta}
}
func (f *deliveryFixture) allowPasswordLogin(t *testing.T) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(deliveryPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	f.users.user.PasswordHash = string(hash)
	f.users.selectEmail = func(email string) (domain.User, error) {
		if email != f.users.user.Email {
			t.Errorf("email = %q", email)
		}
		return f.users.user, nil
	}
}
func deliverySignedToken(t *testing.T, meta domain.JwtAuthMeta, expires time.Time) string {
	t.Helper()
	claims := domain.AccessTokenPayload{
		User:             domain.User{UserID: 42, Email: "buyer@example.com"},
		RegisteredClaims: jwt.RegisteredClaims{Issuer: meta.JwtIssuer, IssuedAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)), ExpiresAt: jwt.NewNumericDate(expires)},
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(meta.JwtSecret)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func assertDeliveryJWT(t *testing.T, raw string, f *deliveryFixture) {
	t.Helper()
	claims := &domain.AccessTokenPayload{}
	token, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) { return f.meta.JwtSecret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(f.meta.JwtIssuer), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		t.Fatalf("invalid access JWT: %v", err)
	}
	if claims.UserID != 42 || claims.Email != f.users.user.Email {
		t.Errorf("unexpected JWT user: %+v", claims.User)
	}
	if claims.PasswordHash != "" {
		t.Error("JWT exposes password hash")
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatal("missing JWT timestamps")
	}
}
func deliveryHash(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
func assertDeliveryCookie(t *testing.T, w *httptest.ResponseRecorder, stored domain.RefreshToken) {
	t.Helper()
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %v, want one refresh cookie", cookies)
	}
	c := cookies[0]
	if c.Name != "refreshToken" || c.Value == "" || c.Path != "/api/v1/auth/refresh" || !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
		t.Errorf("unexpected refresh cookie: %+v", c)
	}
	if c.MaxAge != int(domain.RefreshTokenTTL.Seconds()) || time.Until(c.Expires) < domain.RefreshTokenTTL-time.Minute || time.Until(c.Expires) > domain.RefreshTokenTTL+time.Minute {
		t.Errorf("unexpected cookie lifetime: %+v", c)
	}
	if stored.TokenHash != deliveryHash(c.Value) {
		t.Error("cookie does not match stored token hash")
	}
	if stored.UserID != 42 || stored.DeviceName != defaultDeviceName || stored.IsRevoked {
		t.Errorf("unexpected stored token: %+v", stored)
	}
}
func assertDeliveryEmptyError(t *testing.T, w *httptest.ResponseRecorder, status int) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, status, w.Body.String())
	}
	if w.Body.Len() != 0 {
		t.Errorf("unexpected error body: %s", w.Body.String())
	}
	if len(w.Result().Cookies()) != 0 {
		t.Error("error response sets cookies")
	}
}
func deliveryJSON(t *testing.T, w *httptest.ResponseRecorder, keys ...string) map[string]json.RawMessage {
	t.Helper()
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q", w.Header().Get("Content-Type"))
	}
	var body map[string]json.RawMessage
	decoder := json.NewDecoder(w.Body)
	if err := decoder.Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body) != len(keys) {
		t.Errorf("unexpected response fields: %v", body)
	}
	for _, key := range keys {
		if _, ok := body[key]; !ok {
			t.Errorf("missing field %q", key)
		}
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Errorf("trailing response data: %v", err)
	}
	return body
}
func deliveryJSONString(t *testing.T, body map[string]json.RawMessage, key string) string {
	t.Helper()
	var value string
	if err := json.Unmarshal(body[key], &value); err != nil {
		t.Fatal(err)
	}
	return value
}
func deliveryRequest(path, body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/api/v1/auth/"+path, strings.NewReader(body))
}

type deliveryTrackingBody struct {
	io.Reader
	closed bool
}

func (b *deliveryTrackingBody) Close() error { b.closed = true; return nil }
