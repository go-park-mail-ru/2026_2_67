package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/auth/usecase"
)

type loginRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type loginResponse struct {
	UserID      int64  `json:"userId"`
	AccessToken string `json:"accessToken"`
}

// Login реализует контролер POST /login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	logger := h.logger(r.Context())

	encoder := json.NewEncoder(w)
	accessToken, _ := middleware.ParseAccessTokenPayload(r, h.jwtMeta.JwtSecret)

	body := &loginRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(body)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if body.Email == nil || body.Password == nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	out, err := usecase.Login(
		usecase.LoginInDTO{AccessToken: accessToken,
			Email:      *body.Email,
			Password:   *body.Password,
			DeviceName: deviceNameOf(r)},
		h.tokenRepo, h.userRepo, h.jwtMeta,
	)

	if errors.Is(err, usecase.ErrAlreadyLogined) {
		logger.Warn(err.Error())
		w.Header().Set("Content-Type", "application/json")
		err = encoder.Encode(loginResponse{
			accessToken.UserID, out.AccessToken,
		})
		if err != nil {
			logger.Warn(err.Error())
		}
		return
	}
	if errors.Is(err, domain.ErrEmailNotFound) {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if errors.Is(err, usecase.ErrPasswordInvalid) {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if errors.Is(err, usecase.ErrRefreshTokenNotCreated) {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if errors.Is(err, usecase.ErrAccessTokenNotCreated) {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !errors.Is(err, nil) {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	setRefreshTokenCookie(w, out.RefreshToken)

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(loginResponse{
		out.UserID, out.AccessToken,
	})
	if err != nil {
		logger.Error(err.Error())
	}
}
