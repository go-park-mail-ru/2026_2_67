// Package auth для аутентификации/авторизации пользователей
package auth

import "bmstuozon/backend/internal/storage"

type AuthHandler struct {
	jwtSecret []byte
	jwtIssuer string
	storage   storage.Storage
}

func NewAuthHandler(jwtSecret []byte, jwtIssuer string, storage storage.Storage) *AuthHandler {
	return &AuthHandler{jwtSecret, jwtIssuer, storage}
}
