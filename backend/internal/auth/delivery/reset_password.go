package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/auth/usecase"
)

type resetPasswordRequest struct {
	Password    *string `json:"password"`
	NewPassword *string `json:"newPassword"`
}

type resetPasswordErrResponse struct {
	PasswordErrMsg string `json:"passwordErrMessage"`
}

// ResetPassword реализует контроллер POST /api/v1/auth/password и POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	logger := h.logger(r.Context())

	if r.Header.Get("Authorization") == "" {
		logger.Warn("Authorization header missing")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accessToken, err := middleware.ParseAccessTokenPayload(r, h.jwtMeta.JwtSecret)
	if err != nil || accessToken == nil {
		logger.Warn("invalid or expired accessToken")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	body := &resetPasswordRequest{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(body); err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var newPassword string
	if body.NewPassword != nil && *body.NewPassword != "" {
		newPassword = *body.NewPassword
	} else if body.Password != nil && *body.Password != "" {
		newPassword = *body.Password
	} else {
		logger.Warn("password is empty or not provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	out, err := usecase.ResetPassword(usecase.ResetPasswordInDTO{
		UserID:   accessToken.UserID,
		Password: newPassword,
	}, h.userRepo)

	if errors.Is(err, usecase.ErrInvalidPassword) {
		logger.Warn(err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(resetPasswordErrResponse{
			PasswordErrMsg: out.PasswordErrMsg,
		})
		return
	}

	if errors.Is(err, domain.ErrUserIDNotFound) {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
