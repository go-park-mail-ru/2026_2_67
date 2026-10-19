package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

	encoder := json.NewEncoder(w)
	accessToken, _ := middleware.ParseAccessTokenPayload(r, h.jwtMeta.JwtSecret)

	body := &loginRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if body.Email == nil || body.Password == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	out, err := usecase.LoginUsecase(
		usecase.LoginInDTO{AccessToken: accessToken,
			Email:      *body.Email,
			Password:   *body.Password,
			DeviceName: getDeviceNameOf(r)},
		h.tokenRepo, h.userRepo, h.jwtMeta,
	)

	if errors.Is(err, usecase.ErrAlreadyLogined) {
		w.Header().Set("Content-Type", "application/json")
		err = encoder.Encode(loginResponse{
			accessToken.UserID, out.AccessToken,
		})
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	if errors.Is(err, usecase.ErrEmailNotFound) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if errors.Is(err, usecase.ErrPasswordInvalid) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if errors.Is(err, usecase.ErrRefreshTokenNotCreated) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if errors.Is(err, usecase.ErrAccessTokenNotCreated) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	setRefreshTokenCookie(w, out.RefreshToken)

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(loginResponse{
		out.UserID, out.AccessToken,
	})
	if err != nil {
		// для логирования
		fmt.Println(err)
	}
}
