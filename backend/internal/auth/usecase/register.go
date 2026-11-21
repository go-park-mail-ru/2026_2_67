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

func (u *AuthUsecaseImpl) Register(dto RegisterInDTO) (RegisterOutDTO, error) {
	email := strings.TrimSpace(dto.Email)

	emailErr := ValidateEmail(email)
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

	ctx, err := u.userRepo.BeginTx()
	if err != nil {
		return RegisterOutDTO{}, err
	}
	defer u.userRepo.RollbackTx(ctx)

	user, err := u.userRepo.InsertTx(ctx, email, string(passwordHash))
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
			Issuer:    u.jwtMeta.JwtIssuer,
		}, u.jwtMeta.JwtSecret)

	if err != nil {
		return RegisterOutDTO{}, ErrAccessTokenNotCreated
	}

	// создание refreshToken
	refreshTokenRaw := makeRefreshTokenRaw()
	if refreshTokenRaw == "" {
		return RegisterOutDTO{}, ErrRefreshTokenNotCreated
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
		return RegisterOutDTO{}, ErrRefreshTokenNotCreated
	}

	buyerName, _, _ := strings.Cut(user.Email, "@")
	err = u.buyerRepo.InsertTx(ctx, buyerDomain.Buyer{Name: buyerName, UserID: user.UserID})
	if err != nil {
		return RegisterOutDTO{}, ErrDataInvalid
	}

	u.userRepo.CommitTx(ctx)

	return RegisterOutDTO{UserID: user.UserID, AccessToken: accessToken, RefreshToken: refreshTokenRaw}, nil
}
