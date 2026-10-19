package controller

import (
	"net/http"
	"time"
	"vibe_market/backend/internal/auth/domain"
)

const refreshTokenCookieName string = "refreshToken"
const refreshTokenCookiePath string = "/api/v1/auth/refresh"

func setRefreshTokenCookie(w http.ResponseWriter, refreshTokenRaw string) {
	refreshTokenCookie := &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshTokenRaw,
		Path:     refreshTokenCookiePath,
		Expires:  time.Now().Add(domain.RefreshTokenTTL),
		MaxAge:   int(domain.RefreshTokenTTL.Seconds()),
		HttpOnly: true,                    // fetch не может подхватить
		SameSite: http.SameSiteStrictMode, // CSRF
	}

	http.SetCookie(w, refreshTokenCookie)
}
