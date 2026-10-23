package usecase

import (
	"errors"
	"time"
	"vibe_market/backend/internal/auth/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RefreshInDTO struct {
	RefreshToken string
	DeviceName   string
}

type RefreshOutDTO struct {
	AccessToken  string
	RefreshToken string
}

var (
	ErrRefreshTokenInvalid = errors.New("refresh_token невалиден")
)

func RefreshUsecase(dto RefreshInDTO, tokenRepo domain.RefreshTokenRepo, userRepo domain.UserRepo, jwtMeta domain.JwtAuthMeta) (RefreshOutDTO, error) {
	refreshTokenHash, err := bcrypt.GenerateFromPassword([]byte(dto.RefreshToken), bcrypt.DefaultCost)
	if err != nil {
		return RefreshOutDTO{}, ErrRefreshTokenNotCreated
	}
	refreshToken, err := tokenRepo.SelectByHashAndDeviceName(string(refreshTokenHash), dto.DeviceName)
	// токен не существует
	if errors.Is(err, domain.ErrRefreshTokenNotFound) {
		return RefreshOutDTO{}, domain.ErrRefreshTokenNotFound
	}
	// токен не валиден
	if errors.Is(err, domain.ErrRefreshTokenExpired) || errors.Is(err, domain.ErrRefreshTokenRevoked) {
		return RefreshOutDTO{}, ErrRefreshTokenInvalid
	}

	user, err := userRepo.SelectByUserID(refreshToken.UserID)
	// пользователь не найден по ID
	if err != nil {
		return RefreshOutDTO{}, ErrRefreshTokenInvalid
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

	refreshTokenHash, err = bcrypt.GenerateFromPassword([]byte(refreshTokenRaw), bcrypt.DefaultCost)
	if err != nil {
		return RefreshOutDTO{}, ErrAccessTokenNotCreated
	}

	err = tokenRepo.Insert(domain.RefreshToken{
		UserID:     user.UserID,
		TokenHash:  string(refreshTokenHash),
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
