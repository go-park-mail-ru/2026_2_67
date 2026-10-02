package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

// UserAccessTokenKey используется для безопасного хранения данных в r.Context()
const UserAccessTokenKey contextKey = "UserAccessToken"

// UserAccessToken описывает полезную нагрузку (payload) вашего токена
type UserAccessToken struct {
	User
	jwt.RegisteredClaims
}

// AccessTokenMiddleware передаёт в r.Context *CustomClaims, если AccessToken валиден, иначе nil.
// По сути авторизация через accessToken
func AccessTokenMiddleware(secretKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var claims *UserAccessToken

			// 1. Извлекаем заголовок Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				// 2. Проверяем формат "Bearer <token>"
				headerParts := strings.Split(authHeader, " ")
				if len(headerParts) == 2 && strings.ToLower(headerParts[0]) == "bearer" {
					tokenString := headerParts[1]
					accessToken := &UserAccessToken{}

					// 3. Парсим и валидируем токен
					jwt.ParseWithClaims(tokenString, accessToken, func(token *jwt.Token) (any, error) {
						// Проверяем алгоритм подписи
						if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
							return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
						}
						return secretKey, nil
					})
				}
			}

			// 4. Прокидываем пользовательские данные в context запроса
			ctx := context.WithValue(r.Context(), UserAccessTokenKey, claims)

			// 5. Передаём управление следующему обработчику
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
