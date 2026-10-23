package auth

import (
	"net/http"
	"os"
	"vibe_market/backend/internal/auth/controller"
	"vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/common"
)

// WARN: При создании кортежа в domain.UserRepo должен автоматически создаваться кортеж в buyerDomain.BuyerRepo.
// WARN: Создание происходит на уровне базы данных. Да, это ошибка проектировщика :(

func NewServeMux(userRepo domain.UserRepo, tokenRepo domain.RefreshTokenRepo) *http.ServeMux {
	jwtMeta := domain.JwtAuthMeta{JwtIssuer: os.Getenv("JWT_ISSUER"), JwtSecret: []byte(os.Getenv("JWT_SECRET"))}

	authHandler := controller.NewAuthHandler(jwtMeta, userRepo, tokenRepo)

	mux := http.NewServeMux()

	loginHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Login),
	)
	mux.Handle("/login", loginHandler)

	registerHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Register),
	)
	mux.Handle("/register", registerHandler)

	refreshHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Refresh),
	)
	mux.Handle("/refresh", refreshHandler)

	logoutHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Logout),
	)
	mux.Handle("/logout", logoutHandler)

	return mux
}
