package usecase

import (
	"errors"
	"time"
	"vibe_market/backend/internal/auth/domain"

	"github.com/golang-jwt/jwt/v5"
)

func (u *AuthUsecase) Refresh(dto domain.RefreshInDTO) (domain.RefreshOutDTO, error) {
	refreshToken, err := u.tokenRepo.SelectByHashAndDeviceName(hashRefreshToken(dto.RefreshToken), dto.DeviceName)
	// токен не существует
	if errors.Is(err, domain.ErrRefreshTokenNotFound) {
		return domain.RefreshOutDTO{}, domain.ErrRefreshTokenNotFound
	}
	// токен не валиден
	if errors.Is(err, domain.ErrRefreshTokenExpired) || errors.Is(err, domain.ErrRefreshTokenRevoked) {
		return domain.RefreshOutDTO{}, err
	}
	if err != nil {
		return domain.RefreshOutDTO{}, err
	}

	user, err := u.userRepo.SelectByUserID(refreshToken.UserID)
	// пользователь не найден по ID
	if err != nil {
		return domain.RefreshOutDTO{}, domain.ErrUserIDNotFound
	}

	// создание accessToken
	accessToken, err := makeAccessToken(
		domain.AccessTokenPayload{
			UserID:    user.UserID,
			Email:     user.Email,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(domain.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    u.jwtMeta.JwtIssuer,
		}, u.jwtMeta.JwtSecret)

	if err != nil {
		return domain.RefreshOutDTO{}, ErrAccessTokenNotCreated
	}

	refreshTokenRaw := makeRefreshTokenRaw()
	if refreshTokenRaw == "" {
		return domain.RefreshOutDTO{}, ErrRefreshTokenNotCreated
	}

	err = u.tokenRepo.Revoke(refreshToken.UserID, refreshToken.DeviceName)
	if err != nil {
		return domain.RefreshOutDTO{}, err
	}
	err = u.tokenRepo.Insert(domain.RefreshToken{
		UserID:     user.UserID,
		TokenHash:  hashRefreshToken(refreshTokenRaw),
		DeviceName: refreshToken.DeviceName,
		IsRevoked:  false,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(domain.RefreshTokenTTL),
	})
	if err != nil {
		return domain.RefreshOutDTO{}, ErrRefreshTokenNotCreated
	}

	return domain.RefreshOutDTO{accessToken, refreshTokenRaw}, nil
}
