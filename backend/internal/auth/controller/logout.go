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
	accessToken, err := middleware.ParseAccessTokenPayload(r, h.jwtMeta.JwtSecret)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = usecase.LogoutUsecase(*accessToken, getDeviceNameOf(r), h.tokenRepo)
	if errors.Is(err, usecase.ErrRefreshTokenNotRemoved) {
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
