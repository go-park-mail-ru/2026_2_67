package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/golang-jwt/jwt/v5"
)

func getSignedAccessToken(accessToken UserAccessToken, jwtSecret []byte) (string, error) {
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

// getHashOf возвращает sha256 хеш для str
func getHashOf(str string) string {
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:])
}
