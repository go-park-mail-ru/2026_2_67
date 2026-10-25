package delivery

import (
	"encoding/json"
	"errors"
	"net/http"

	"vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/auth/usecase"
)

type registerRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type unregisteredResponse struct {
	EmailErrMsg    string `json:"emailErrMessage"`
	PasswordErrMsg string `json:"passwordErrMessage"`
}

type registerResponse struct {
	UserID      int64  `json:"userId"`
	AccessToken string `json:"accessToken"`
}

// Register реализует контролер POST /register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	logger := h.logger(r.Context())

	accessToken, _ := middleware.ParseAccessTokenPayload(r, h.jwtMeta.JwtSecret)
	if accessToken != nil {
		logger.Warn(usecase.ErrAlreadyLogined.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body := &registerRequest{}
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(body)
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if body.Email == nil || body.Password == nil {
		logger.Warn("body.Email or body.Password not sent")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	out, err := usecase.Register(
		usecase.RegisterInDTO{Email: *body.Email, Password: *body.Password, DeviceName: deviceNameOf(r)}, h.userRepo,
		h.tokenRepo, h.buyerRepo, h.jwtMeta)

	if errors.Is(err, usecase.ErrDataInvalid) {
		logger.Warn(err.Error())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		err := encoder.Encode(unregisteredResponse{
			out.EmailErrMsg, out.PasswordErrMsg,
		})
		if err != nil {
			logger.Error(err.Error())
		}
		return
	}
	if errors.Is(err, domain.ErrEmailAlreadyExists) {
		logger.Warn(err.Error())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		err := encoder.Encode(unregisteredResponse{
			out.EmailErrMsg, "",
		})
		if err != nil {
			logger.Error(err.Error())
		}
		return
	}
	if errors.Is(err, usecase.ErrAccessTokenNotCreated) {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if errors.Is(err, usecase.ErrRefreshTokenNotCreated) {
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
	err = encoder.Encode(registerResponse{
		out.UserID,
		out.AccessToken,
	})
	if err != nil {
		logger.Error(err.Error())
		return
	}
}
