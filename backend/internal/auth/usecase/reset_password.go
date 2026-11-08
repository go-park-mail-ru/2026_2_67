package usecase

import (
	"errors"
	"vibe_market/backend/internal/auth/domain"

	"golang.org/x/crypto/bcrypt"
)

type ResetPasswordInDTO struct {
	UserID   int64
	Password string
}

type ResetPasswordOutDTO struct {
	PasswordErrMsg string
}

var (
	ErrInvalidPassword = errors.New("пароль невалиден")
)

func ResetPassword(dto ResetPasswordInDTO, userRepo domain.UserRepo) (ResetPasswordOutDTO, error) {
	if err := validatePassword(dto.Password); err != nil {
		return ResetPasswordOutDTO{PasswordErrMsg: err.Error()}, ErrInvalidPassword
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return ResetPasswordOutDTO{}, err
	}

	err = userRepo.UpdatePasswordByUserID(dto.UserID, string(passwordHash))
	if err != nil {
		if errors.Is(err, domain.ErrUserIDNotFound) {
			return ResetPasswordOutDTO{}, domain.ErrUserIDNotFound
		}
		return ResetPasswordOutDTO{}, err
	}

	return ResetPasswordOutDTO{}, nil
}
