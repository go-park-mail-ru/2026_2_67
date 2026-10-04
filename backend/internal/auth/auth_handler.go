// Package auth для аутентификации/авторизации пользователей
package auth

import "vibe_market/backend/internal/storage"

type AuthHandler struct {
	jwtSecret []byte
	jwtIssuer string
	storage   storage.Storage
}

func NewAuthHandler(jwtSecret []byte, jwtIssuer string, storage storage.Storage) *AuthHandler {
	return &AuthHandler{jwtSecret, jwtIssuer, storage}
}
