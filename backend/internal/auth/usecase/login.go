package usecase

import (
	"errors"
	"strings"
	"time"
	"vibe_market/backend/internal/auth/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginInDTO struct {
	AccessToken *domain.AccessTokenPayload
	Email       string
	Password    string
	DeviceName  string
}

type LoginOutDTO struct {
	UserID       int64
	AccessToken  string
	RefreshToken string
}

var (
	ErrEmailNotFound          = errors.New("пользователь по email не найден")
	ErrPasswordInvalid        = errors.New("неверный пароль")
	ErrRefreshTokenNotCreated = errors.New("refreshToken не создан")
	ErrAccessTokenNotCreated  = errors.New("accessToken не создан")
	ErrAlreadyLogined         = errors.New("уже залогинин")
)

func LoginUsecase(dto LoginInDTO, tokenRepo domain.RefreshTokenRepo, userRepo domain.UserRepo, jwtMeta domain.JwtAuthMeta) (LoginOutDTO, error) {
	if dto.AccessToken != nil {
		accessTokenResponse, err := makeAccessToken(*dto.AccessToken, jwtMeta.JwtSecret)
		if err != nil {
			return LoginOutDTO{}, ErrAccessTokenNotCreated
		}
		return LoginOutDTO{dto.AccessToken.UserID, accessTokenResponse, ""}, ErrAlreadyLogined
	}

	email := strings.TrimSpace(dto.Email)

	user, ok := userRepo.SelectByEmail(email)
	// пользователь не найден (email не найден )
	if !ok {
		return LoginOutDTO{}, ErrEmailNotFound
	}
	// пароль в бд не совпадает с введенным
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password)) != nil {
		return LoginOutDTO{}, ErrPasswordInvalid
	}

	// создание accessToken
	accessTokenResponse, err := makeAccessToken(
		domain.AccessTokenPayload{
			UserID:    user.UserID,
			Email:     user.Email,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(domain.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    jwtMeta.JwtIssuer,
		}, jwtMeta.JwtSecret)

	if err != nil {
		return LoginOutDTO{}, ErrAccessTokenNotCreated
	}

	// создание refreshToken
	refreshTokenRaw := makeRefreshTokenRaw()
	if refreshTokenRaw == "" {
		return LoginOutDTO{}, ErrRefreshTokenNotCreated
	}
	refreshTokenHash, err := bcrypt.GenerateFromPassword([]byte(refreshTokenRaw), bcrypt.DefaultCost)
	if err != nil {
		return LoginOutDTO{}, ErrRefreshTokenNotCreated
	}

	err = tokenRepo.Insert(domain.RefreshToken{
		UserID:     user.UserID,
		TokenHash:  string(refreshTokenHash),
		DeviceName: dto.DeviceName,
		IsRevoked:  false,
		ExpiresAt:  time.Now().Add(domain.RefreshTokenTTL),
		CreatedAt:  time.Now(),
	})

	if err != nil {
		return LoginOutDTO{}, ErrRefreshTokenNotCreated
	}

	return LoginOutDTO{user.UserID, accessTokenResponse, refreshTokenRaw}, nil
}
