package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/usecase"
)

type refreshResponse struct {
	AcccessToken string `json:"accessToken"`
}

// Refresh реализует контролер POST /refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	refreshTokenRaw, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	out, err := usecase.Refresh(usecase.RefreshInDTO{
		RefreshToken: refreshTokenRaw.Value, DeviceName: deviceNameOf(r),
	}, h.tokenRepo, h.userRepo, h.jwtMeta)

	if errors.Is(err, domain.ErrRefreshTokenNotFound) {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if errors.Is(err, usecase.ErrRefreshTokenInvalid) {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if errors.Is(err, usecase.ErrAccessTokenNotCreated) {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if errors.Is(err, usecase.ErrRefreshTokenNotCreated) {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	setRefreshTokenCookie(w, out.RefreshToken)
	encoder := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(refreshResponse{out.AccessToken})
	if err != nil {
		logger.Error(err.Error())
		return
	}
}
