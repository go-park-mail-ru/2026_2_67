package usecase

import (
	"errors"
	"time"
	"vibe_market/backend/internal/auth/domain"

	"github.com/golang-jwt/jwt/v5"
)

type RefreshInDTO struct {
	RefreshToken string
	DeviceName   string
}

type RefreshOutDTO struct {
	AccessToken  string
	RefreshToken string
}

func Refresh(dto RefreshInDTO, tokenRepo domain.RefreshTokenRepo, userRepo domain.UserRepo, jwtMeta domain.JwtAuthMeta) (RefreshOutDTO, error) {
	refreshToken, err := tokenRepo.SelectByHashAndDeviceName(hashRefreshToken(dto.RefreshToken), dto.DeviceName)
	// токен не существует
	if errors.Is(err, domain.ErrRefreshTokenNotFound) {
		return RefreshOutDTO{}, domain.ErrRefreshTokenNotFound
	}
	// токен не валиден
	if errors.Is(err, domain.ErrRefreshTokenExpired) || errors.Is(err, domain.ErrRefreshTokenRevoked) {
		return RefreshOutDTO{}, err
	}
	if err != nil {
		return RefreshOutDTO{}, err
	}

	user, err := userRepo.SelectByUserID(refreshToken.UserID)
	// пользователь не найден по ID
	if err != nil {
		return RefreshOutDTO{}, domain.ErrUserIDNotFound
	}

	// создание accessToken
	accessToken, err := makeAccessToken(
		domain.AccessTokenPayload{
			UserID:    user.UserID,
			Email:     user.Email,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(domain.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    jwtMeta.JwtIssuer,
		}, jwtMeta.JwtSecret)

	if err != nil {
		return RefreshOutDTO{}, ErrAccessTokenNotCreated
	}

	refreshTokenRaw := makeRefreshTokenRaw()
	if refreshTokenRaw == "" {
		return RefreshOutDTO{}, ErrRefreshTokenNotCreated
	}

	err = tokenRepo.Revoke(refreshToken.UserID, refreshToken.DeviceName)
	if err != nil {
		return RefreshOutDTO{}, err
	}
	err = tokenRepo.Insert(domain.RefreshToken{
		UserID:     user.UserID,
		TokenHash:  hashRefreshToken(refreshTokenRaw),
		DeviceName: refreshToken.DeviceName,
		IsRevoked:  false,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(domain.RefreshTokenTTL),
	})
	if err != nil {
		return RefreshOutDTO{}, ErrRefreshTokenNotCreated
	}

	return RefreshOutDTO{accessToken, refreshTokenRaw}, nil
}
