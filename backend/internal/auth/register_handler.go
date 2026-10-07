package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"vibe_market/backend/internal/storage"

	"github.com/golang-jwt/jwt/v5"
)

type registerRequest struct {
	Login    *string `json:"login"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type unregisteredResponse struct {
	LoginErrMessage    string `json:"loginErrMessage"`
	EmailErrMessage    string `json:"emailErrMessage"`
	PasswordErrMessage string `json:"passwordErrMessage"`
}

type registerResponse struct {
	UserID      int64  `json:"userId"`
	AccessToken string `json:"accessToken"`
}

// Register реализует контролер POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	accessToken, _ := parseAccessTokenPayload(r, h.jwtSecret)
	if accessToken != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	requestBody := &registerRequest{}
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if requestBody.Login == nil || requestBody.Email == nil || requestBody.Password == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	login := strings.TrimSpace(*requestBody.Login)
	email := strings.TrimSpace(*requestBody.Email)

	loginErr := validateLogin(login)
	emailErr := validateEmail(email)
	passwordErr := validatePassword(*requestBody.Password)

	if emailErr != nil || passwordErr != nil || loginErr != nil {
		emailErrMsg := ""
		if emailErr != nil {
			emailErrMsg = emailErr.Error()
		}
		passwordErrMsg := ""
		if passwordErr != nil {
			passwordErrMsg = passwordErr.Error()
		}
		loginErrMsg := ""
		if loginErr != nil {
			loginErrMsg = loginErr.Error()
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		err := encoder.Encode(unregisteredResponse{
			loginErrMsg, emailErrMsg, passwordErrMsg,
		})
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	user, err := h.storage.InsertUser(login, email, *requestBody.Password)
	if err != nil {
		emailErrMsg := ""
		if errors.Is(err, storage.ErrUserEmailAlreadyExists) {
			emailErrMsg = storage.ErrUserEmailAlreadyExists.Error()
		}
		loginErrMsg := ""
		if errors.Is(err, storage.ErrUserLoginAlreadyExists) {
			loginErrMsg = storage.ErrUserLoginAlreadyExists.Error()
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		err := encoder.Encode(unregisteredResponse{
			loginErrMsg, emailErrMsg, "",
		})
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	// создание accessToken
	accessTokenResponse, err := makeAccessToken(
		AccessTokenPayload{
			UserID:    user.UserID,
			Login:     login,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    h.jwtIssuer,
		}, h.jwtSecret)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// создание refreshToken
	refreshTokenRaw := makeRefreshTokenRaw()
	if refreshTokenRaw == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	refreshTokenHash := makeHashOf(refreshTokenRaw)

	err = h.storage.InsertRefreshToken(storage.RefreshToken{
		UserID:    user.UserID,
		TokenHash: refreshTokenHash,
		IsRevoked: false,
		ExpiresAt: time.Now().Add(refreshTokenTTL),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// отправка RegisterResponse
	setRefreshTokenCookie(w, refreshTokenRaw)

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(registerResponse{
		user.UserID,
		accessTokenResponse,
	})
	if err != nil {
		fmt.Println(err)
	}
}
