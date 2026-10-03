// Package auth для аутентификации/авторизации пользователей
package auth

import (
	"net/http"
	"time"
)

type AuthHandler struct {
	jwtSecret []byte
	jwtIssuer string
	storage   Storage
}

func NewAuthHandler(jwtSecret []byte, jwtIssuer string, storage Storage) *AuthHandler {
	return &AuthHandler{jwtSecret, jwtIssuer, storage}
}

func setRefreshTokenCookie(w http.ResponseWriter, refreshTokenRaw string) {
	ttl := 7 * 24 * time.Hour

	refreshTokenCookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshTokenRaw,
		Path:     "/auth/refresh",
		Expires:  time.Now().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,                    // Защита от JS (XSS)
		SameSite: http.SameSiteStrictMode, // Защита от CSRF
	}

	http.SetCookie(w, refreshTokenCookie)
}
