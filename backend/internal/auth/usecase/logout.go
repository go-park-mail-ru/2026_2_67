package usecase

import (
	"errors"
	"vibe_market/backend/internal/auth/domain"
)

var (
	ErrRefreshTokenNotRemoved = errors.New("refreshToken не удален")
)

func LogoutUsecase(accessToken domain.AccessTokenPayload, deviceName string, tokenRepo domain.RefreshTokenRepo) error {
	if !tokenRepo.Revoke(accessToken.UserID, deviceName) {
		return ErrRefreshTokenNotRemoved
	}

	return nil
}
