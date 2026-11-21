package usecase

import (
	"errors"
	"strings"
	"time"
	"vibe_market/backend/internal/auth/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordInvalid        = errors.New("неверный пароль")
	ErrRefreshTokenNotCreated = errors.New("refreshToken не создан")
	ErrAccessTokenNotCreated  = errors.New("accessToken не создан")
	ErrAlreadyLogined         = errors.New("уже залогинин")
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

func (u *AuthUsecaseImpl) Login(dto LoginInDTO) (LoginOutDTO, error) {
	if dto.AccessToken != nil {
		accessTokenResponse, err := makeAccessToken(*dto.AccessToken, u.jwtMeta.JwtSecret)
		if err != nil {
			return LoginOutDTO{}, ErrAccessTokenNotCreated
		}
		return LoginOutDTO{dto.AccessToken.UserID, accessTokenResponse, ""}, ErrAlreadyLogined
	}

	email := strings.TrimSpace(dto.Email)

	user, err := u.userRepo.SelectByEmail(email)
	// пользователь не найден (email не найден )
	if errors.Is(err, domain.ErrEmailNotFound) {
		return LoginOutDTO{}, domain.ErrEmailNotFound
	}
	if err != nil {
		return LoginOutDTO{}, err
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
			Issuer:    u.jwtMeta.JwtIssuer,
		}, u.jwtMeta.JwtSecret)

	if err != nil {
		return LoginOutDTO{}, ErrAccessTokenNotCreated
	}

	// создание refreshToken
	refreshTokenRaw := makeRefreshTokenRaw()
	if refreshTokenRaw == "" {
		return LoginOutDTO{}, ErrRefreshTokenNotCreated
	}

	err = u.tokenRepo.Insert(domain.RefreshToken{
		UserID:     user.UserID,
		TokenHash:  hashRefreshToken(refreshTokenRaw),
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
