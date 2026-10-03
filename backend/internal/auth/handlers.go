// Package auth для аутентификации/авторизации пользователей
package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	jwtSecret []byte
	storage   Storage
}

func MakeAuthHandler(jwtSecret []byte, storage Storage) AuthHandler {
	return AuthHandler{jwtSecret, storage}
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterStatusUnauthorizedResponse struct {
	ErrMessage string `json:"errMessage"`
}

type RegisterResponse struct {
	UserID       int64  `json:"userId"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	requestBody := &RegisterRequest{}
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = IsValidPassword(requestBody.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")

		err = encoder.Encode(RegisterStatusUnauthorizedResponse{
			ErrMessage: err.Error(),
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := h.storage.InsertUser(requestBody.Login, requestBody.Email, requestBody.Password)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// создание accessToken
	accessToken, err := makeAccessToken(user, RoleBuyer, h.jwtSecret)
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
	refreshTokenHash := getHashOf(refreshTokenRaw)

	err = h.storage.InsertRefreshToken(RefreshToken{
		UserID:    user.UserID,
		TokenHash: refreshTokenHash,
		IsRevoked: false,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(RegisterResponse{
		UserID:       user.UserID,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenRaw,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type LoginRequest struct {
	LoginOrEmail string `json:"loginOrEmail"`
	Password     string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// LoginHandler реализует роутер POST /auth/login
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	encoder := json.NewEncoder(w)
	accessToken, ok := r.Context().Value(UserAccessTokenKey).(*UserAccessToken)

	// внешней код не передал AccessToken
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// пользователь авторизован
	if accessToken != nil {
		accessTokenSigned, err := getSignedAccessToken(*accessToken, h.jwtSecret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = encoder.Encode(LoginResponse{
			AccessToken:  accessTokenSigned,
			RefreshToken: h.storage.SelectRefreshTokenOf(accessToken.UserID),
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	requestBody := &LoginRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, ok := h.storage.SelectUserByLoginOrEmail(requestBody.LoginOrEmail)
	// пользователь не найден
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// не совпадают пароли
	if bcrypt.CompareHashAndPassword([]byte(user.passwordHash), []byte(requestBody.Password)) != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// создание refreshToken
	refreshTokenRaw := makeRefreshTokenRaw()
	if refreshTokenRaw == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	refreshTokenHash := getHashOf(refreshTokenRaw)

	err = h.storage.InsertRefreshToken(RefreshToken{
		UserID:    user.UserID,
		TokenHash: refreshTokenHash,
		IsRevoked: false,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// создание accessToken
	accessTokenSigned, err := makeAccessToken(user, RoleBuyer, h.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(LoginResponse{
		AccessToken:  accessTokenSigned,
		RefreshToken: refreshTokenRaw,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func makeAccessToken(user User, role RoleType, secret []byte) (string, error) {
	return getSignedAccessToken(
		UserAccessToken{
			UserID:    user.UserID,
			Login:     user.Login,
			Role:      role,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ozon", // TODO: поменять
		}, secret)
}
