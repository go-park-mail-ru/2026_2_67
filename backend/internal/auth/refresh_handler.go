package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type refreshResponse struct {
	AcccessToken string `json:"accessToken"`
}

// Refresh реализует роутер POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshTokenRaw, ok := r.Cookie("refreshToken")
	if ok != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	refreshTokenHash := makeHashOf(refreshTokenRaw.Value)

	refreshToken, found := h.storage.SelectRefreshTokenByHash(refreshTokenHash)
	// токен существует
	if !found {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// токен валиден
	if refreshToken.IsRevoked || refreshToken.ExpiresAt.Before(time.Now()) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	user, found := h.storage.SelectUserByID(refreshToken.UserID)
	// пользователь не найден по ID
	if !found {
		w.WriteHeader(http.StatusForbidden)
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

	encoder := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(refreshResponse{
		accessTokenResponse,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
