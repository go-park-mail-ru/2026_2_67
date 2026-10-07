package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"vibe_market/backend/internal/storage"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	LoginOrEmail string `json:"loginOrEmail"`
	Password     string `json:"password"`
}

type loginResponse struct {
	UserID      int64  `json:"userId"`
	AccessToken string `json:"accessToken"`
}

// Login реализует роутер POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	encoder := json.NewEncoder(w)
	accessToken, _ := parseAccessTokenPayload(r, h.jwtSecret)

	// пользователь уже авторизован
	if accessToken != nil {
		accessTokenResponse, err := makeAccessToken(*accessToken, h.jwtSecret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = encoder.Encode(loginResponse{
			accessToken.UserID, accessTokenResponse,
		})
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	requestBody := &loginRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	loginOrEmail := strings.TrimSpace(requestBody.LoginOrEmail)

	user, ok := h.storage.SelectUserByLoginOrEmail(loginOrEmail)
	// пользователь не найден (login или email не найден в бд)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// пароль в бд не совпадает с введенным
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(requestBody.Password)) != nil {
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

	// создание accessToken
	accessTokenResponse, err := makeAccessToken(
		AccessTokenPayload{
			UserID:    user.UserID,
			Login:     user.Login,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    h.jwtIssuer,
		}, h.jwtSecret)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// отправка LoginResponse
	setRefreshTokenCookie(w, refreshTokenRaw)

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(loginResponse{
		user.UserID, accessTokenResponse,
	})
	if err != nil {
		fmt.Println(err)
	}
}
