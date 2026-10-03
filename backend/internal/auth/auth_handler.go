// Package auth для аутентификации/авторизации пользователей
package auth

type AuthHandler struct {
	jwtSecret []byte
	jwtIssuer string
	storage   Storage
}

func NewAuthHandler(jwtSecret []byte, jwtIssuer string, storage Storage) *AuthHandler {
	return &AuthHandler{jwtSecret, jwtIssuer, storage}
}
