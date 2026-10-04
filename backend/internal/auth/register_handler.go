package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type registerRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerUnauthorizedResponse struct {
	EmailErrMessage    string `json:"emailErrMessage"`
	PasswordErrMessage string `json:"passwordErrMessage"`
}

type registerConflictResponse struct {
	LoginErrMessage string `json:"loginErrMessage"`
	EmailErrMessage string `json:"emailErrMessage"`
}

type registerResponse struct {
	UserID      int64  `json:"userId"`
	AccessToken string `json:"accessToken"`
}

// Register реализует роутер POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	accessToken, _ := ParseAccessTokenPayload(r, h.jwtSecret)
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

	emailErr := validateEmail(requestBody.Email)
	passwordErr := validatePassword(requestBody.Password)

	if emailErr != nil || passwordErr != nil {
		emailErrMsg := ""
		if emailErr != nil {
			emailErrMsg = emailErr.Error()
		}
		passwordErrMsg := ""
		if passwordErr != nil {
			passwordErrMsg = passwordErr.Error()
		}

		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		err := encoder.Encode(registerUnauthorizedResponse{
			emailErrMsg, passwordErrMsg,
		})
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	user, err := h.storage.InsertUser(requestBody.Login, requestBody.Email, requestBody.Password)
	if err != nil {
		emailErrMsg := ""
		if errors.Is(err, errEmailAlreadyExists) {
			emailErrMsg = errEmailAlreadyExists.Error()
		}
		loginErrMsg := ""
		if errors.Is(err, errLoginAlreadyExists) {
			loginErrMsg = errLoginAlreadyExists.Error()
		}

		w.WriteHeader(http.StatusConflict)
		w.Header().Set("Content-Type", "application/json")
		err := encoder.Encode(registerConflictResponse{
			loginErrMsg, emailErrMsg,
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
			Login:     user.Login,
			Role:      user.Role,
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

	err = h.storage.InsertRefreshToken(RefreshToken{
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
