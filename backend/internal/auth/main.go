package auth

import (
	"net/http"
	"os"
	"vibe_market/backend/internal/auth/controller"
	"vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/repo"
	"vibe_market/backend/internal/common"
)

func NewServeMux() *http.ServeMux {
	jwtMeta := domain.JwtAuthMeta{JwtIssuer: os.Getenv("JWT_ISSUER"), JwtSecret: []byte(os.Getenv("JWT_SECRET"))}
	userRepo := repo.NewUserInMemoryDB()
	refreshTokenRepo := repo.NewRefreshTokenInMemoryDB()

	authHandler := controller.NewAuthHandler(jwtMeta, userRepo, refreshTokenRepo)

	mux := http.NewServeMux()

	// POST /api/v1/auth/login
	loginHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Login),
	)
	mux.Handle("/login", loginHandler)

	// POST /api/v1/auth/register
	registerHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Register),
	)
	mux.Handle("/register", registerHandler)

	// POST /api/v1/auth/refresh
	refreshHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Refresh),
	)
	mux.Handle("/refresh", refreshHandler)

	// POST /api/v1/auth/logout
	logoutHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Logout),
	)
	mux.Handle("/logout", logoutHandler)

	return mux
}
