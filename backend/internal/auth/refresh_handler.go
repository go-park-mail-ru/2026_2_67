package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"vibe_market/backend/internal/storage"

	"github.com/golang-jwt/jwt/v5"
)

type refreshResponse struct {
	AcccessToken string `json:"accessToken"`
}

// Refresh реализует роутер POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshTokenRaw, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	refreshTokenHash := makeHashOf(refreshTokenRaw.Value)

	refreshToken, ok := h.storage.SelectRefreshTokenByHash(refreshTokenHash)
	// токен не существует
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// токен не валиден
	if refreshToken.IsRevoked || refreshToken.ExpiresAt.Before(time.Now()) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	user, ok := h.storage.SelectUserByID(refreshToken.UserID)
	// пользователь не найден по ID
	if !ok {
		w.WriteHeader(http.StatusForbidden)
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
	// обновление refreshToken и получение нового accessToken для user
	h.storage.DropRefreshToken(user.UserID)

	newRefreshTokenRaw := makeRefreshTokenRaw()
	if newRefreshTokenRaw == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	setRefreshTokenCookie(w, newRefreshTokenRaw)
	encoder := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(refreshResponse{
		accessTokenResponse,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	refreshTokenHash = makeHashOf(newRefreshTokenRaw)

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
}
