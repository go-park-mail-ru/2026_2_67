package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

// AccessTokenPayloadKey используется для безопасного хранения данных в r.Context()
const AccessTokenPayloadKey contextKey = "UserAccessToken"

// AccessTokenPayload описывает полезную нагрузку (payload) JWT токена
type AccessTokenPayload struct {
	User
	jwt.RegisteredClaims
}

func ParseAccessTokenPayload(r *http.Request, secretKey []byte) (*AccessTokenPayload, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("HTTP-заголовок 'Authorization' отсутствует")
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
		return nil, errors.New("неправильный формат 'Authorization'")
	}

	tokenString := headerParts[1]
	accessTokenPayload := &AccessTokenPayload{}

	token, err := jwt.ParseWithClaims(tokenString, accessTokenPayload, func(token *jwt.Token) (any, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неверный метод подписи: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token error")
	}

	return accessTokenPayload, nil
}

// AccessTokenPayloadMiddleware передаёт в r.Context *AccessTokenPayload, если accessToken валиден, иначе nil.
// По сути авторизация через accessToken.
func AccessTokenPayloadMiddleware(secretKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accessTokenPayload, _ := ParseAccessTokenPayload(r, secretKey)
			ctx := context.WithValue(r.Context(), AccessTokenPayloadKey, accessTokenPayload)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
