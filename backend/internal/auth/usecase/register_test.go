package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
	"vibe_market/backend/internal/auth/domain"
	buyerDomain "vibe_market/backend/internal/buyers/domain"

	"golang.org/x/crypto/bcrypt"
)

func TestRegisterValidation(t *testing.T) {
	for _, tc := range []struct {
		name, email, password string
		emailErr, passwordErr error
	}{
		{"empty email", " \t", authTestPassword, errEmailEmpty, nil},
		{"invalid email", "not-an-email", authTestPassword, domain.ErrInvalidEmail, nil},
		{"empty password", "buyer@example.com", "", nil, errPasswordTooShort},
		{"short password", "buyer@example.com", "Aa1!", nil, errPasswordTooShort},
		{"no uppercase", "buyer@example.com", "password123!", nil, errNoUpper},
		{"no lowercase", "buyer@example.com", "PASSWORD123!", nil, errNoLower},
		{"no number", "buyer@example.com", "Password!", nil, errNoNumber},
		{"no special", "buyer@example.com", "Password123", nil, errNoSpecial},
		{"space", "buyer@example.com", "Password 123!", nil, errHasSpace},
		{"unicode whitespace", "buyer@example.com", "Password\u00a0123!", nil, errHasSpace},
		{"both invalid", "bad", "short", domain.ErrInvalidEmail, errPasswordTooShort},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := Register(RegisterInDTO{Email: tc.email, Password: tc.password}, &authUserStub{}, &authTokenStub{}, &authBuyerStub{}, authTestMeta)
			want := RegisterOutDTO{}
			if tc.emailErr != nil {
				want.EmailErrMsg = tc.emailErr.Error()
			}
			if tc.passwordErr != nil {
				want.PasswordErrMsg = tc.passwordErr.Error()
			}
			if !errors.Is(err, ErrInvalidData) || out != want {
				t.Fatalf("Register = %+v, %v; want %+v, ErrInvalidData", out, err, want)
			}
		})
	}
}

func TestRegisterBcryptLengthLimit(t *testing.T) {
	// Valid according to the validator, but beyond bcrypt's 72-byte limit.
	out, err := Register(RegisterInDTO{Email: "buyer@example.com", Password: "Aa1!" + strings.Repeat("x", 69)}, &authUserStub{}, &authTokenStub{}, &authBuyerStub{}, authTestMeta)
	if !errors.Is(err, bcrypt.ErrPasswordTooLong) || out != (RegisterOutDTO{}) {
		t.Fatalf("Register = %+v, %v", out, err)
	}
}

func TestRegisterTransactions(t *testing.T) {
	dbErr := errors.New("database failure")
	for _, tc := range []struct {
		name                                          string
		beginErr, insertErr, tokenErr, buyerErr, want error
		emailMessage                                  string
		calls                                         []string
	}{
		{name: "success", calls: []string{"begin", "user", "token", "buyer", "commit", "rollback"}},
		{name: "begin failure", beginErr: dbErr, want: dbErr, calls: []string{"begin"}},
		{name: "duplicate email", insertErr: domain.ErrEmailAlreadyExists, want: domain.ErrEmailAlreadyExists, emailMessage: domain.ErrEmailAlreadyExists.Error(), calls: []string{"begin", "user", "rollback"}},
		{name: "wrapped duplicate email", insertErr: fmt.Errorf("insert: %w", domain.ErrEmailAlreadyExists), want: domain.ErrEmailAlreadyExists, emailMessage: domain.ErrEmailAlreadyExists.Error(), calls: []string{"begin", "user", "rollback"}},
		// Characterization: unrelated insertion errors are reported as duplicate email.
		{name: "user insert infrastructure error masked", insertErr: dbErr, want: domain.ErrEmailAlreadyExists, calls: []string{"begin", "user", "rollback"}},
		{name: "token insert failure", tokenErr: dbErr, want: ErrRefreshTokenNotCreated, calls: []string{"begin", "user", "token", "rollback"}},
		// Token insertion is outside the user/buyer transaction and has no compensation.
		{name: "buyer insert failure after token persisted", buyerErr: dbErr, want: ErrInvalidData, calls: []string{"begin", "user", "token", "buyer", "rollback"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var calls []string
			checkCtx := func(got context.Context) {
				t.Helper()
				if got != ctx {
					t.Fatal("transaction context was not propagated")
				}
			}
			user := domain.User{UserID: 42, Email: "buyer@example.com"}
			users := &authUserStub{
				begin: func() (context.Context, error) { calls = append(calls, "begin"); return ctx, tc.beginErr },
				insert: func(got context.Context, email, hash string) (domain.User, error) {
					calls = append(calls, "user")
					checkCtx(got)
					if email != user.Email {
						t.Fatalf("email = %q", email)
					}
					if hash == authTestPassword {
						t.Fatal("plaintext password stored")
					}
					if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(authTestPassword)); err != nil {
						t.Fatalf("invalid bcrypt hash: %v", err)
					}
					if cost, err := bcrypt.Cost([]byte(hash)); err != nil || cost != bcrypt.DefaultCost {
						t.Fatalf("bcrypt cost = %d, %v", cost, err)
					}
					user.PasswordHash = hash
					return user, tc.insertErr
				},
				commit:   func(got context.Context) { calls = append(calls, "commit"); checkCtx(got) },
				rollback: func(got context.Context) { calls = append(calls, "rollback"); checkCtx(got) },
			}
			var stored domain.RefreshToken
			tokens := &authTokenStub{insert: func(token domain.RefreshToken) error {
				calls = append(calls, "token")
				stored = token
				return tc.tokenErr
			}}
			buyers := &authBuyerStub{insert: func(got context.Context, buyer buyerDomain.Buyer) error {
				calls = append(calls, "buyer")
				checkCtx(got)
				if buyer.UserID != 42 || buyer.Name != "buyer" || buyer.Surname != nil || buyer.Telephone != nil {
					t.Fatalf("buyer = %+v", buyer)
				}
				return tc.buyerErr
			}}
			start := time.Now()
			out, err := Register(RegisterInDTO{Email: " \t" + user.Email + "\n", Password: authTestPassword, DeviceName: "browser"}, users, tokens, buyers, authTestMeta)
			end := time.Now()
			if !errors.Is(err, tc.want) {
				t.Fatalf("Register error = %v, want %v", err, tc.want)
			}
			assertAuthCalls(t, calls, tc.calls)
			if tc.want != nil {
				if out != (RegisterOutDTO{EmailErrMsg: tc.emailMessage}) {
					t.Fatalf("unexpected failure output: %+v", out)
				}
				return
			}
			if out.UserID != 42 || out.EmailErrMsg != "" || out.PasswordErrMsg != "" {
				t.Fatalf("Register output = %+v", out)
			}
			assertAuthJWT(t, out.AccessToken, user, start, end)
			assertAuthRefresh(t, out.RefreshToken, stored, 42, "browser", start, end)
		})
	}
}
