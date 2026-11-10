package usecase

import (
	"errors"
	"fmt"
	"testing"
	"time"
	"vibe_market/backend/internal/auth/domain"
)

func TestRefreshRotation(t *testing.T) {
	user := domain.User{UserID: 42, Email: "buyer@example.com", PasswordHash: "must-not-leak"}
	old := domain.RefreshToken{UserID: 42, DeviceName: "browser", TokenHash: authRefreshHash("old-token"), ExpiresAt: time.Now().Add(time.Hour)}
	var stored domain.RefreshToken
	var calls []string
	tokens := &authTokenStub{
		selectToken: func(hash, device string) (domain.RefreshToken, error) {
			calls = append(calls, "select")
			if hash != old.TokenHash || device != "browser" {
				t.Fatalf("lookup(%q,%q)", hash, device)
			}
			if old.IsRevoked {
				return domain.RefreshToken{}, domain.ErrRefreshTokenRevoked
			}
			return old, nil
		},
		revoke: func(id int64, device string) error {
			calls = append(calls, "revoke")
			if id != 42 || device != "browser" {
				t.Fatalf("revoke(%d,%q)", id, device)
			}
			old.IsRevoked = true
			return nil
		},
		insert: func(token domain.RefreshToken) error {
			calls = append(calls, "insert")
			if !old.IsRevoked {
				t.Fatal("old token must be revoked before insert")
			}
			stored = token
			return nil
		},
	}
	users := &authUserStub{byID: func(id int64) (domain.User, error) {
		calls = append(calls, "user")
		if id != 42 {
			t.Fatalf("user lookup = %d", id)
		}
		return user, nil
	}}
	start := time.Now()
	out, err := Refresh(RefreshInDTO{RefreshToken: "old-token", DeviceName: "browser"}, tokens, users, authTestMeta)
	end := time.Now()
	if err != nil {
		t.Fatal(err)
	}
	assertAuthCalls(t, calls, []string{"select", "user", "revoke", "insert"})
	assertAuthJWT(t, out.AccessToken, user, start, end)
	assertAuthRefresh(t, out.RefreshToken, stored, 42, "browser", start, end)
	if out.RefreshToken == "old-token" || stored.TokenHash == old.TokenHash {
		t.Fatal("refresh token was not rotated")
	}
	calls = nil
	replay, err := Refresh(RefreshInDTO{RefreshToken: "old-token", DeviceName: "browser"}, tokens, users, authTestMeta)
	if !errors.Is(err, domain.ErrRefreshTokenRevoked) || replay != (RefreshOutDTO{}) {
		t.Fatalf("replay = %+v, %v", replay, err)
	}
	assertAuthCalls(t, calls, []string{"select"})
}

func TestRefreshFailures(t *testing.T) {
	dbErr := errors.New("database failure")
	for _, tc := range []struct {
		name                                           string
		selectErr, userErr, revokeErr, insertErr, want error
		calls                                          []string
	}{
		{name: "missing", selectErr: domain.ErrRefreshTokenNotFound, want: domain.ErrRefreshTokenNotFound, calls: []string{"select"}},
		{name: "wrapped missing", selectErr: fmt.Errorf("select: %w", domain.ErrRefreshTokenNotFound), want: domain.ErrRefreshTokenNotFound, calls: []string{"select"}},
		{name: "expired", selectErr: domain.ErrRefreshTokenExpired, want: domain.ErrRefreshTokenExpired, calls: []string{"select"}},
		{name: "revoked", selectErr: domain.ErrRefreshTokenRevoked, want: domain.ErrRefreshTokenRevoked, calls: []string{"select"}},
		{name: "lookup failure", selectErr: dbErr, want: dbErr, calls: []string{"select"}},
		{name: "user missing", userErr: domain.ErrUserIDNotFound, want: domain.ErrUserIDNotFound, calls: []string{"select", "user"}},
		// Characterization: an infrastructure error is currently masked as a missing user.
		{name: "user database failure is masked", userErr: dbErr, want: domain.ErrUserIDNotFound, calls: []string{"select", "user"}},
		{name: "revoke failure", revokeErr: dbErr, want: dbErr, calls: []string{"select", "user", "revoke"}},
		// Characterization: the old session is already revoked when insertion fails.
		{name: "insert failure after revoke", insertErr: dbErr, want: ErrRefreshTokenNotCreated, calls: []string{"select", "user", "revoke", "insert"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			tokens := &authTokenStub{
				selectToken: func(hash, device string) (domain.RefreshToken, error) {
					calls = append(calls, "select")
					if hash != authRefreshHash("raw") || device != "request-device" {
						t.Fatalf("lookup(%q,%q)", hash, device)
					}
					return domain.RefreshToken{UserID: 42, DeviceName: "stored-device"}, tc.selectErr
				},
				revoke: func(id int64, device string) error {
					calls = append(calls, "revoke")
					if id != 42 || device != "stored-device" {
						t.Fatalf("revoke(%d,%q)", id, device)
					}
					return tc.revokeErr
				},
				insert: func(token domain.RefreshToken) error {
					calls = append(calls, "insert")
					if token.DeviceName != "stored-device" {
						t.Fatalf("insert device = %q", token.DeviceName)
					}
					return tc.insertErr
				},
			}
			users := &authUserStub{byID: func(id int64) (domain.User, error) {
				calls = append(calls, "user")
				if id != 42 {
					t.Fatalf("user lookup = %d", id)
				}
				return domain.User{UserID: 42}, tc.userErr
			}}
			out, err := Refresh(RefreshInDTO{RefreshToken: "raw", DeviceName: "request-device"}, tokens, users, authTestMeta)
			if !errors.Is(err, tc.want) || out != (RefreshOutDTO{}) {
				t.Fatalf("Refresh = %+v, %v; want %v", out, err, tc.want)
			}
			assertAuthCalls(t, calls, tc.calls)
		})
	}
}
