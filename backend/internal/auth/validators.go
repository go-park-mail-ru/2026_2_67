package auth

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	errPasswordTooShort = errors.New("пароль должен быть длиной не менее 8 символов")
	errNoUpper          = errors.New("пароль должен содержать хотя бы одну заглавную букву")
	errNoLower          = errors.New("пароль должен содержать хотя бы одну строчную букву")
	errNoNumber         = errors.New("пароль должен содержать хотя бы одну цифру")
	errNoSpecial        = errors.New("пароль должен содержать хотя бы один спецсимвол")
	errHasSpace         = errors.New("пароль не должен содержать пробельные символы")
)

// validatePassword проверяет пароль на соответствие базовым стандартам безопасности.
func validatePassword(password string) error {
	runes := []rune(password)
	length := len(runes)

	if length < 8 {
		return errPasswordTooShort
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

var (
	errEmptyLogin   = errors.New("логин не может быть пустым")
	errLoginLetters = errors.New("логин может содержать только латинские буквы, цифры, дефис и подчеркивание")
)

func validateLogin(login string) error {
	login = strings.TrimSpace(login)
	if login == "" {
		return errEmptyLogin
	}

	loginRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !loginRegex.MatchString(login) {
		return errLoginLetters
	}

	return nil
}

var (
	errEmailEmpty      = errors.New("email не может быть пустым")
	errEmailInvalidFmt = errors.New("некорректный формат email")
)

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errEmailEmpty
	}
	emailRegex := regexp.MustCompile("^[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$")
	if !emailRegex.MatchString(email) {
		return errEmailInvalidFmt
	}

	return nil
}
