package auth

import (
	"errors"
	"unicode"
)

var (
	errPasswordTooShort = errors.New("пароль должен быть длиной не менее 8 символов")
	errPasswordTooLong  = errors.New("пароль должен быть длиной не более 64 символов")
	errNoUpper          = errors.New("пароль должен содержать хотя бы одну заглавную букву")
	errNoLower          = errors.New("пароль должен содержать хотя бы одну строчную букву")
	errNoNumber         = errors.New("пароль должен содержать хотя бы одну цифру")
	errNoSpecial        = errors.New("пароль должен содержать хотя бы один спецсимвол")
	errHasSpace         = errors.New("пароль не должен содержать пробельные символы")
)

// isValidPassword проверяет пароль на соответствие базовым стандартам безопасности.
func isValidPassword(password string) error {
	runes := []rune(password)
	length := len(runes)

	if length < 8 {
		return errPasswordTooShort
	}
	if length > 64 {
		return errPasswordTooLong
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
			return errHasSpace
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
		return errNoUpper
	}
	if !hasLower {
		return errNoLower
	}
	if !hasNumber {
		return errNoNumber
	}
	if !hasSpecial {
		return errNoSpecial
	}

	return nil
}
