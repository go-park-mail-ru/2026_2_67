package auth

import (
	"net/http"
	"time"
)

const refreshTokenCookieName string = "refreshToken"
const refreshTokenTTL time.Duration = 7 * 24 * time.Hour

func setRefreshTokenCookie(w http.ResponseWriter, refreshTokenRaw string) {
	refreshTokenCookie := &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshTokenRaw,
		Path:     "/auth/refresh",
		Expires:  time.Now().Add(refreshTokenTTL),
		MaxAge:   int(refreshTokenTTL.Seconds()),
		HttpOnly: true,                    // Защита от JS (XSS)
		SameSite: http.SameSiteStrictMode, // Защита от CSRF
	}

	http.SetCookie(w, refreshTokenCookie)
}
