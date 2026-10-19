package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"vibe_market/backend/internal/auth/usecase"
)

type refreshResponse struct {
	AcccessToken string `json:"accessToken"`
}

// Refresh реализует контролер POST /refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshTokenRaw, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	out, err := usecase.RefreshUsecase(usecase.RefreshInDTO{
		RefreshToken: refreshTokenRaw.Value, DeviceName: getDeviceNameOf(r),
	}, h.tokenRepo, h.userRepo, h.jwtMeta)

	if errors.Is(err, usecase.ErrRefreshTokenNotFound) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if errors.Is(err, usecase.ErrRefreshTokenInvalid) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if errors.Is(err, usecase.ErrAccessTokenNotCreated) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if errors.Is(err, usecase.ErrRefreshTokenNotCreated) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	setRefreshTokenCookie(w, out.RefreshToken)
	encoder := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(refreshResponse{out.AccessToken})
	if err != nil {
		fmt.Println(err)
		return
	}
}
