package auth

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
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

// validatePassword проверяет пароль на соответствие базовым стандартам безопасности.
func validatePassword(password string) error {
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
	ErrEmailEmpty      = errors.New("email не может быть пустым")
	ErrEmailInvalidFmt = errors.New("некорректный формат email")
	ErrEmailExtraData  = errors.New("email содержит лишние символы или имя")
	ErrEmailMissingTLD = errors.New("доменная часть должна содержать домен верхнего уровня (например, .com, .ru)")
)

func validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return ErrEmailEmpty
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEmailInvalidFmt, err)
	}

	if addr.Address != email {
		return ErrEmailExtraData
	}

	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 || !strings.Contains(parts[1], ".") {
		return ErrEmailMissingTLD
	}

	return nil
}
