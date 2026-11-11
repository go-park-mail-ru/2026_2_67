package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"testing"
	"time"
	"vibe_market/backend/internal/auth/domain"
	buyerDomain "vibe_market/backend/internal/buyers/domain"

	"github.com/golang-jwt/jwt/v5"
)

// Embedded interfaces deliberately panic on unexpected dependency calls.
type authUserStub struct {
	domain.UserRepo
	begin    func() (context.Context, error)
	insert   func(context.Context, string, string) (domain.User, error)
	commit   func(context.Context)
	rollback func(context.Context)
	byEmail  func(string) (domain.User, error)
	byID     func(int64) (domain.User, error)
}

func (s *authUserStub) BeginTx() (context.Context, error) { return s.begin() }
func (s *authUserStub) InsertTx(ctx context.Context, email, hash string) (domain.User, error) {
	return s.insert(ctx, email, hash)
}
func (s *authUserStub) CommitTx(ctx context.Context)                    { s.commit(ctx) }
func (s *authUserStub) RollbackTx(ctx context.Context)                  { s.rollback(ctx) }
func (s *authUserStub) SelectByEmail(email string) (domain.User, error) { return s.byEmail(email) }
func (s *authUserStub) SelectByUserID(id int64) (domain.User, error)    { return s.byID(id) }

type authTokenStub struct {
	insert      func(domain.RefreshToken) error
	revoke      func(int64, string) error
	selectToken func(string, string) (domain.RefreshToken, error)
}

func (s *authTokenStub) Insert(token domain.RefreshToken) error { return s.insert(token) }
func (s *authTokenStub) Revoke(id int64, device string) error   { return s.revoke(id, device) }
func (s *authTokenStub) SelectByHashAndDeviceName(hash, device string) (domain.RefreshToken, error) {
	return s.selectToken(hash, device)
}

type authBuyerStub struct {
	buyerDomain.BuyerRepo
	insert func(context.Context, buyerDomain.Buyer) error
}

func (s *authBuyerStub) InsertTx(ctx context.Context, buyer buyerDomain.Buyer) error {
	return s.insert(ctx, buyer)
}

var authTestMeta = domain.JwtAuthMeta{JwtIssuer: "auth-usecase-test", JwtSecret: []byte("test-only-signing-secret")}

const authTestPassword = "ValidPassword123!"

func assertAuthCalls(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dependency calls = %v, want %v", got, want)
	}
}

func assertAuthJWT(t *testing.T, raw string, user domain.User, start, end time.Time) {
	t.Helper()
	claims := &domain.AccessTokenPayload{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) { return authTestMeta.JwtSecret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(authTestMeta.JwtIssuer), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid {
		t.Fatalf("invalid signed access token: %v", err)
	}
	if claims.UserID != user.UserID || claims.Email != user.Email || claims.PasswordHash != "" {
		t.Fatalf("unexpected JWT user claims: %+v", claims.User)
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatal("missing JWT timestamps")
	}
	assertAuthTime(t, claims.IssuedAt.Time, start.Truncate(time.Second), end.Truncate(time.Second))
	assertAuthTime(t, claims.ExpiresAt.Time, start.Add(domain.AccessTokenTTL).Truncate(time.Second), end.Add(domain.AccessTokenTTL).Truncate(time.Second))
	if _, err := jwt.Parse(raw, func(*jwt.Token) (any, error) { return []byte("wrong-secret"), nil }); err == nil {
		t.Fatal("JWT accepted with wrong signing secret")
	}
}

func assertAuthTime(t *testing.T, got, earliest, latest time.Time) {
	t.Helper()
	if got.Before(earliest) || got.After(latest) {
		t.Fatalf("time %v outside [%v, %v]", got, earliest, latest)
	}
}

func authRefreshHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func assertAuthRefresh(t *testing.T, raw string, stored domain.RefreshToken, userID int64, device string, start, end time.Time) {
	t.Helper()
	bytes, err := hex.DecodeString(raw)
	if err != nil || len(bytes) != 32 {
		t.Fatalf("refresh token must be 32 random bytes in hex: %q (%v)", raw, err)
	}
	if stored.UserID != userID || stored.DeviceName != device || stored.IsRevoked || stored.TokenHash != authRefreshHash(raw) || stored.TokenHash == raw {
		t.Fatalf("unexpected stored refresh token: %+v", stored)
	}
	assertAuthTime(t, stored.CreatedAt, start, end)
	assertAuthTime(t, stored.ExpiresAt, start.Add(domain.RefreshTokenTTL), end.Add(domain.RefreshTokenTTL))
}

var _ domain.UserRepo = (*authUserStub)(nil)
var _ domain.RefreshTokenRepo = (*authTokenStub)(nil)
var _ buyerDomain.BuyerRepo = (*authBuyerStub)(nil)
