package domain

import "errors"

type LoginInDTO struct {
	AccessToken *AccessTokenPayload
	Email       string
	Password    string
	DeviceName  string
}

type LoginOutDTO struct {
	UserID       int64
	AccessToken  string
	RefreshToken string
}

type RefreshInDTO struct {
	RefreshToken string
	DeviceName   string
}

type RefreshOutDTO struct {
	AccessToken  string
	RefreshToken string
}

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

type AuthUsecase interface {
	Login(dto LoginInDTO) (LoginOutDTO, error)
	Logout(accessToken AccessTokenPayload, deviceName string, tokenRepo RefreshTokenRepo) error
	Refresh(dto RefreshInDTO) (RefreshOutDTO, error)
	Register(dto RegisterInDTO) (RegisterOutDTO, error)
}

var (
	ErrDataInvalid = errors.New("email и password невалидны")
)
