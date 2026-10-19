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

var (
	ErrRefreshTokenNotFound = errors.New("refreshToken не найден")
	ErrRefreshTokenInvalid  = errors.New("refreshToken невалиден")
)

func RefreshUsecase(dto RefreshInDTO, tokenRepo domain.RefreshTokenRepo, userRepo domain.UserRepo, jwtMeta domain.JwtAuthMeta) (RefreshOutDTO, error) {
	refreshToken, ok := tokenRepo.SelectByHashAndDeviceName(makeHashOf(dto.RefreshToken), dto.DeviceName)
	// токен не существует
	if !ok {
		return RefreshOutDTO{}, ErrRefreshTokenNotFound
	}
	// токен не валиден
	if refreshToken.IsRevoked || refreshToken.ExpiresAt.Before(time.Now()) {
		return RefreshOutDTO{}, ErrRefreshTokenInvalid
	}

	user, ok := userRepo.SelectByUserID(refreshToken.UserID)
	// пользователь не найден по ID
	if !ok {
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

	refreshTokenHash := makeHashOf(refreshTokenRaw)

	err = tokenRepo.Insert(domain.RefreshToken{
		UserID:     user.UserID,
		TokenHash:  refreshTokenHash,
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
