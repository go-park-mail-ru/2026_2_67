package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"vibe_market/backend/internal/auth/domain"

	"github.com/golang-jwt/jwt/v5"
)

// makeAccessToken подписывает accessToken (для клиента)
func makeAccessToken(accessToken domain.AccessTokenPayload, jwtSecret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessToken)
	signedToken, err := token.SignedString(jwtSecret)
	return signedToken, err
}

// makeRefreshTokenRaw создает сырой токен (для клиента)
func makeRefreshTokenRaw() string {
	// 1. Генерируем 32 случайных байта
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}
