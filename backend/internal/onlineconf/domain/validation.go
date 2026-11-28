package domain

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrKeyInvalid   = errors.New("некорректный key: допустимы только a-z, 0-9, '.', '_', '-', длина 1..64")
	ErrScopeInvalid = errors.New("некорректный scope: допустимы только a-z, 0-9, '.', '_', '-', длина 1..64")
)

var identifierPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9._-]{0,62}[a-z0-9])?$`)

// ValidateIdentifier проверяет key и scope на допустимые символы (только латиница в нижнем регистре, цифры, ., _, -).
func ValidateIdentifier(value string) bool {
	if !identifierPattern.MatchString(value) {
		return false
	}
	return !strings.Contains(value, "..")
}
