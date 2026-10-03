package auth

import (
	"encoding/json"
	"net/http"
	"time"

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

// Login реализует роутер POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	encoder := json.NewEncoder(w)
	accessToken, _ := ParseAccessTokenPayload(r, h.jwtSecret)

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
			w.WriteHeader(http.StatusInternalServerError)
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
	setRefreshTokenCookie(w, refreshTokenRaw)

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(loginResponse{
		user.UserID, accessTokenResponse,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
