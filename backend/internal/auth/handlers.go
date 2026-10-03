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
	jwtIssuer string
	storage   Storage
}

func MakeAuthHandler(jwtSecret []byte, jwtIssuer string, storage Storage) AuthHandler {
	return AuthHandler{jwtSecret, jwtIssuer, storage}
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
	UserID      int64  `json:"userId"`
	AccessToken string `json:"accessToken"`
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
			err.Error(),
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// TODO: добавить проверку, что requestBody.email это email
	user, err := h.storage.InsertUser(requestBody.Login, requestBody.Email, requestBody.Password)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// создание accessToken
	accessToken, err := makeAccessToken(
		AccessTokenPayload{
			UserID:    user.UserID,
			Login:     user.Login,
			Role:      user.Role,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
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
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// отправка RegisterResponse
	setRefreshTokenCookie(&w, refreshTokenRaw)

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(RegisterResponse{
		user.UserID,
		accessToken,
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
	UserID      int64  `json:"userId"`
	AccessToken string `json:"accessToken"`
}

// LoginHandler реализует роутер POST /auth/login
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	encoder := json.NewEncoder(w)
	accessToken, ok := r.Context().Value(AccessTokenPayloadKey).(*AccessTokenPayload)

	// внешний код не сделал AccessTokenMiddleware
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// пользователь уже авторизован
	if accessToken != nil {
		accessTokenResponse, err := makeAccessToken(*accessToken, h.jwtSecret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = encoder.Encode(LoginResponse{
			accessToken.UserID, accessTokenResponse,
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
	// пользователь не найден (login или email не найден в бд)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// пароль в бд не совпадает с введенным
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
	refreshTokenHash := makeHashOf(refreshTokenRaw)

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
	accessTokenResponse, err := makeAccessToken(
		AccessTokenPayload{
			UserID:    user.UserID,
			Login:     user.Login,
			Role:      user.Role,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    h.jwtIssuer,
		}, h.jwtSecret)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// отправка LoginResponse
	setRefreshTokenCookie(&w, refreshTokenRaw)

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(LoginResponse{
		user.UserID, accessTokenResponse,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func setRefreshTokenCookie(w *http.ResponseWriter, refreshTokenRaw string) {
	ttl := 7 * 24 * time.Hour

	refreshTokenCookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshTokenRaw,
		Path:     "/auth/refresh",
		Expires:  time.Now().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,                    // Защита от JS (XSS)
		SameSite: http.SameSiteStrictMode, // Защита от CSRF
	}

	http.SetCookie(*w, refreshTokenCookie)
}
