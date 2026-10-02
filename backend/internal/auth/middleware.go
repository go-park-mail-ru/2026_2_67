package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

// UserClaimsKey используется для безопасного хранения данных в r.Context()
const UserClaimsKey contextKey = "UserClaims"

// CustomClaims описывает полезную нагрузку (payload) вашего токена
type CustomClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAccessTokenMiddleware передаёт в r.Context *CustomClaims, если AccessToken валиден, иначе nil
func JWTAccessTokenMiddleware(secretKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var claims *CustomClaims

			// 1. Извлекаем заголовок Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				// 2. Проверяем формат "Bearer <token>"
				headerParts := strings.Split(authHeader, " ")
				if len(headerParts) == 2 && strings.ToLower(headerParts[0]) == "bearer" {
					tokenString := headerParts[1]
					claims := &CustomClaims{}

					// 3. Парсим и валидируем токен
					token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
						// Проверяем алгоритм подписи
						if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
							return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
						}
						return secretKey, nil
					})

					if err != nil || !token.Valid {
						claims = nil
					}
				}
			}

			// 4. Прокидываем пользовательские данные в context запроса
			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)

			// 5. Передаём управление следующему обработчику
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
