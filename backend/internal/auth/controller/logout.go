package controller

import (
	"errors"
	"net/http"
	"time"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/auth/usecase"
)

// Logout реализует контролер POST /logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	accessToken, err := middleware.ParseAccessTokenPayload(r, h.jwtMeta.JwtSecret)

	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = usecase.Logout(*accessToken, deviceNameOf(r), h.tokenRepo)
	if errors.Is(err, usecase.ErrRefreshTokenNotRemoved) {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refreshTokenCookie := &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "revoked",
		Path:     refreshTokenCookiePath,
		Expires:  time.Now().Add(-time.Minute),
		MaxAge:   int(time.Now().Add(-time.Minute).Second()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode}

	http.SetCookie(w, refreshTokenCookie)
}
