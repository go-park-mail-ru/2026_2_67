package usecase

import (
	"crypto/rand"
	"crypto/sha256"
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

func hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// makeRefreshTokenRaw создает сырой токен (для клиента)
func makeRefreshTokenRaw() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}
