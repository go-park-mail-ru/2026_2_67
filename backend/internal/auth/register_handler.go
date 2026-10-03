package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type registerRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerStatusUnauthorizedResponse struct {
	ErrMessage string `json:"errMessage"`
}

type registerResponse struct {
	UserID      int64  `json:"userId"`
	AccessToken string `json:"accessToken"`
}

// Register реализует роутер POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	accessToken, ok := r.Context().Value(AccessTokenPayloadKey).(*AccessTokenPayload)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if accessToken != nil {
		w.WriteHeader(http.StatusConflict)
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

	err = isValidPassword(requestBody.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")

		err = encoder.Encode(registerStatusUnauthorizedResponse{
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
	setRefreshTokenCookie(w, refreshTokenRaw)

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(registerResponse{
		user.UserID,
		accessTokenResponse,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
