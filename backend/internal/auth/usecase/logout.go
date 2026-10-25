package usecase

import (
	"errors"
	"vibe_market/backend/internal/auth/domain"
)

var (
	ErrRefreshTokenNotRemoved = errors.New("refreshToken не удален")
)

func Logout(accessToken domain.AccessTokenPayload, deviceName string, tokenRepo domain.RefreshTokenRepo) error {
	if tokenRepo.Revoke(accessToken.UserID, deviceName) != nil {
		return ErrRefreshTokenNotRemoved
	}

	return nil
}
