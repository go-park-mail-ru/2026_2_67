package usecase

// TODO: Регистрация пользователя должна быть в рамках одной транзакции.
// TODO: Следует понять как эту транзакцию реализовать.
// TODO: то есть, если в момент создания User что-то произойдёт, то User должен удалиться.

import (
	"errors"
	"strings"
	"time"
	"vibe_market/backend/internal/auth/domain"
	buyerDomain "vibe_market/backend/internal/buyers/domain"

	"github.com/golang-jwt/jwt/v5"
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

func RegisterUsecase(dto RegisterInDTO, userRepo domain.UserRepo, buyerRepo buyerDomain.BuyerRepo, tokenRepo domain.RefreshTokenRepo, jwtMeta domain.JwtAuthMeta) (RegisterOutDTO, error) {
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

	user, err := userRepo.Insert(dto.Email, dto.Password)
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
		userRepo.Truncate(user.UserID)
		return RegisterOutDTO{}, ErrAccessTokenNotCreated
	}

	// создание refreshToken
	refreshTokenRaw := makeRefreshTokenRaw()
	if refreshTokenRaw == "" {
		userRepo.Truncate(user.UserID)
		return RegisterOutDTO{}, ErrRefreshTokenNotCreated
	}
	refreshTokenHash := makeHashOf(refreshTokenRaw)

	err = tokenRepo.Insert(domain.RefreshToken{
		UserID:     user.UserID,
		TokenHash:  refreshTokenHash,
		DeviceName: dto.DeviceName,
		IsRevoked:  false,
		ExpiresAt:  time.Now().Add(domain.RefreshTokenTTL),
		CreatedAt:  time.Now(),
	})

	if err != nil {
		userRepo.Truncate(user.UserID)
		return RegisterOutDTO{}, ErrRefreshTokenNotCreated
	}

	tmpName, _, _ := strings.Cut(user.Email, "@")
	buyerRepo.Insert(buyerDomain.Buyer{User: user, Name: tmpName})

	return RegisterOutDTO{UserID: user.UserID, AccessToken: accessToken, RefreshToken: refreshTokenRaw}, nil
}
