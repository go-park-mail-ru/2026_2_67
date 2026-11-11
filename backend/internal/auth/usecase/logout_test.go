package usecase

import (
	"errors"
	"testing"
	"vibe_market/backend/internal/auth/domain"
)

func TestLogout(t *testing.T) {
	for _, tc := range []struct {
		name          string
		repoErr, want error
	}{
		{name: "success"},
		{name: "repository failure", repoErr: errors.New("database failure"), want: ErrRefreshTokenNotRemoved},
		{name: "token not found", repoErr: domain.ErrRefreshTokenNotFound, want: ErrRefreshTokenNotRemoved},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			tokens := &authTokenStub{revoke: func(id int64, device string) error {
				calls++
				if id != 42 || device != "browser" {
					t.Fatalf("Revoke(%d, %q)", id, device)
				}
				return tc.repoErr
			}}
			payload := domain.AccessTokenPayload{}
			payload.UserID = 42
			if err := Logout(payload, "browser", tokens); !errors.Is(err, tc.want) {
				t.Fatalf("Logout error = %v, want %v", err, tc.want)
			}
			if calls != 1 {
				t.Fatalf("Revoke called %d times", calls)
			}
		})
	}
}
