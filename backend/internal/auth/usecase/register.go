package usecase

import (
	"errors"
	"strings"
	"time"
	"vibe_market/backend/internal/auth/domain"
	buyerDomain "vibe_market/backend/internal/buyers/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInDTO struct {
	Email      string
	Password   string
	DeviceName string
}

type RegisterOutDTO struct {
	UserID         int64
	AccessToken    string
	RefreshToken   string
	EmailErrMsg    string
	PasswordErrMsg string
}

var (
	ErrDataInvalid = errors.New("email и password невалидны")
)

// TODO: Регистрация пользователя должна быть в рамках одной транзакции.
// TODO: Следует понять как эту транзакцию реализовать.
// TODO: то есть, если в момент создания User что-то произойдёт, то User должен удалиться.

// WARN: 3 запроса к БД это больно :(

func Register(dto RegisterInDTO, userRepo domain.UserRepo, tokenRepo domain.RefreshTokenRepo,
	buyerRepo buyerDomain.BuyerRepo, jwtMeta domain.JwtAuthMeta) (RegisterOutDTO, error) {
	email := strings.TrimSpace(dto.Email)

	emailErr := validateEmail(email)
	passwordErr := validatePassword(dto.Password)

	if emailErr != nil || passwordErr != nil {
		emailErrMsg := ""
		if emailErr != nil {
			emailErrMsg = emailErr.Error()
		}
		passwordErrMsg := ""
		if passwordErr != nil {
			passwordErrMsg = passwordErr.Error()
		}

		return RegisterOutDTO{EmailErrMsg: emailErrMsg, PasswordErrMsg: passwordErrMsg}, ErrDataInvalid
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterOutDTO{}, err
	}
	user, err := userRepo.Insert(dto.Email, string(passwordHash))
	if err != nil {
		emailErrMsg := ""
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			emailErrMsg = domain.ErrEmailAlreadyExists.Error()
		}

		return RegisterOutDTO{EmailErrMsg: emailErrMsg}, domain.ErrEmailAlreadyExists
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
		userRepo.DeleteByUserID(user.UserID)
		return RegisterOutDTO{}, ErrAccessTokenNotCreated
	}

	// создание refreshToken
	refreshTokenRaw := makeRefreshTokenRaw()
	if refreshTokenRaw == "" {
		userRepo.DeleteByUserID(user.UserID)
		return RegisterOutDTO{}, ErrRefreshTokenNotCreated
	}
	refreshTokenHash, err := bcrypt.GenerateFromPassword([]byte(refreshTokenRaw), bcrypt.DefaultCost)
	if err != nil {
		return RegisterOutDTO{}, ErrRefreshTokenNotCreated
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
		userRepo.DeleteByUserID(user.UserID)
		return RegisterOutDTO{}, ErrRefreshTokenNotCreated
	}

	buyerName, _, _ := strings.Cut(user.Email, "@")
	err = buyerRepo.Insert(buyerDomain.Buyer{Name: buyerName, UserID: user.UserID})
	if err != nil {
		userRepo.DeleteByUserID(user.UserID)
		return RegisterOutDTO{}, ErrDataInvalid
	}

	return RegisterOutDTO{UserID: user.UserID, AccessToken: accessToken, RefreshToken: refreshTokenRaw}, nil
}
