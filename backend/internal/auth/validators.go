package auth

import (
	"errors"
	"fmt"
	"net/mail"
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

// validateEmail проверяет, является ли строка корректным email-адресом.
func validateEmail(email string) error {
	// 1. Базовая проверка на пустую строку
	if strings.TrimSpace(email) == "" {
		return errors.New("email не может быть пустым")
	}

	// 2. Валидация структуры по RFC 5322 через net/mail
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("некорректный формат email: %w", err)
	}

	// 3. Запрет имен с адресной частью в кавычках (например, "John Doe" <john@example.com>)
	if addr.Address != email {
		return errors.New("email содержит лишние символы или имя")
	}

	// 4. Дополнительная проверка: наличие точки в доменной части (защита от "user@localhost")
	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 || !strings.Contains(parts[1], ".") {
		return errors.New("доменная часть должна содержать домен верхнего уровня (например, .com, .ru)")
	}

	return nil
}
