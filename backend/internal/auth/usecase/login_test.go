package usecase

import (
	"errors"
	"fmt"
	"testing"
	"time"
	"vibe_market/backend/internal/auth/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginSuccess(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte(authTestPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user := domain.User{UserID: 42, Email: "buyer@example.com", PasswordHash: string(hash)}
	var calls []string
	users := &authUserStub{byEmail: func(email string) (domain.User, error) {
		calls = append(calls, "select")
		if email != user.Email {
			t.Fatalf("email = %q", email)
		}
		return user, nil
	}}
	var stored domain.RefreshToken
	tokens := &authTokenStub{insert: func(token domain.RefreshToken) error { calls = append(calls, "insert"); stored = token; return nil }}
	start := time.Now()
	out, err := Login(LoginInDTO{Email: " \t" + user.Email + "\n", Password: authTestPassword, DeviceName: "browser"}, tokens, users, authTestMeta)
	end := time.Now()
	if err != nil || out.UserID != user.UserID {
		t.Fatalf("Login = %+v, %v", out, err)
	}
	assertAuthCalls(t, calls, []string{"select", "insert"})
	assertAuthJWT(t, out.AccessToken, user, start, end)
	assertAuthRefresh(t, out.RefreshToken, stored, user.UserID, "browser", start, end)
}

func TestLoginFailures(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte(authTestPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name, password, hash       string
		selectErr, insertErr, want error
		insert                     bool
	}{
		{name: "email missing", selectErr: domain.ErrEmailNotFound, want: domain.ErrEmailNotFound},
		{name: "wrapped email missing", selectErr: fmt.Errorf("select: %w", domain.ErrEmailNotFound), want: domain.ErrEmailNotFound},
		{name: "database failure", selectErr: dbErr, want: dbErr},
		{name: "wrong password", password: "WrongPassword123!", hash: string(hash), want: ErrPasswordInvalid},
		{name: "empty password", hash: string(hash), want: ErrPasswordInvalid},
		{name: "password whitespace is significant", password: " " + authTestPassword, hash: string(hash), want: ErrPasswordInvalid},
		{name: "corrupt bcrypt hash", password: authTestPassword, hash: "not-bcrypt", want: ErrPasswordInvalid},
		{name: "refresh persistence", password: authTestPassword, hash: string(hash), insertErr: dbErr, want: ErrRefreshTokenNotCreated, insert: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			users := &authUserStub{byEmail: func(string) (domain.User, error) {
				calls = append(calls, "select")
				return domain.User{UserID: 42, PasswordHash: tc.hash}, tc.selectErr
			}}
			tokens := &authTokenStub{insert: func(domain.RefreshToken) error { calls = append(calls, "insert"); return tc.insertErr }}
			out, err := Login(LoginInDTO{Email: "buyer@example.com", Password: tc.password}, tokens, users, authTestMeta)
			if !errors.Is(err, tc.want) || out != (LoginOutDTO{}) {
				t.Fatalf("Login = %+v, %v; want empty output, %v", out, err, tc.want)
			}
			want := []string{"select"}
			if tc.insert {
				want = append(want, "insert")
			}
			assertAuthCalls(t, calls, want)
		})
	}
}

func TestLoginAlreadyAuthenticated(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	payload := domain.AccessTokenPayload{}
	payload.UserID = 17
	payload.Email = "existing@example.com"
	payload.Issuer = authTestMeta.JwtIssuer
	payload.IssuedAt = jwt.NewNumericDate(now)
	payload.ExpiresAt = jwt.NewNumericDate(now.Add(domain.AccessTokenTTL))
	out, err := Login(LoginInDTO{AccessToken: &payload, Email: "ignored", Password: "ignored"}, &authTokenStub{}, &authUserStub{}, authTestMeta)
	if !errors.Is(err, ErrAlreadyLogined) || out.UserID != 17 || out.RefreshToken != "" {
		t.Fatalf("Login = %+v, %v", out, err)
	}
	assertAuthJWT(t, out.AccessToken, payload.User, now, now)
}
