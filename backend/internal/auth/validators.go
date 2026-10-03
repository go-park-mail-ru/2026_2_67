package auth

import (
	"errors"
	"unicode"
)

var (
	ErrPasswordTooShort = errors.New("пароль должен быть длиной не менее 8 символов")
	ErrPasswordTooLong  = errors.New("пароль должен быть длиной не более 64 символов")
	ErrNoUpper          = errors.New("пароль должен содержать хотя бы одну заглавную букву")
	ErrNoLower          = errors.New("пароль должен содержать хотя бы одну строчную букву")
	ErrNoNumber         = errors.New("пароль должен содержать хотя бы одну цифру")
	ErrNoSpecial        = errors.New("пароль должен содержать хотя бы один спецсимвол")
	ErrHasSpace         = errors.New("пароль не должен содержать пробельные символы")
)

// IsValidPassword проверяет пароль на соответствие базовым стандартам безопасности.
func IsValidPassword(password string) error {
	runes := []rune(password)
	length := len(runes)

	if length < 8 {
		return ErrPasswordTooShort
	}
	if length > 64 {
		return ErrPasswordTooLong
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, ch := range runes {
		switch {
		case unicode.IsSpace(ch):
			return ErrHasSpace
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsNumber(ch):
			hasNumber = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return ErrNoUpper
	}
	if !hasLower {
		return ErrNoLower
	}
	if !hasNumber {
		return ErrNoNumber
	}
	if !hasSpecial {
		return ErrNoSpecial
	}

	return nil
}
