package auth

import (
	"net/http"
	"time"
)

// Logout реализует контролер POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	accessTokenPayload, err := parseAccessTokenPayload(r, h.jwtSecret)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !h.storage.DropRefreshToken(accessTokenPayload.UserID) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	refreshTokenCookie := &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "revoked",
		Path:     "/api/v1/auth/refresh",
		Expires:  time.Now().Add(-time.Minute),
		MaxAge:   int(time.Now().Add(-time.Minute).Second()),
		HttpOnly: true,                    // Защита от JS (XSS)
		SameSite: http.SameSiteStrictMode, // Защита от CSRF
	}

	http.SetCookie(w, refreshTokenCookie)
}
