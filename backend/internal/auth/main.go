package auth

import (
	"net/http"
	"os"
	"vibe_market/backend/internal/auth/delivery"
	"vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/usecase"
	buyerDomain "vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/common"

	"go.uber.org/zap"
)

func NewServeMux(userRepo domain.UserRepo, tokenRepo domain.RefreshTokenRepo, buyerRepo buyerDomain.BuyerRepo, baseLogger *zap.Logger) *http.ServeMux {
	jwtMeta := domain.JwtAuthMeta{JwtIssuer: os.Getenv("JWT_ISSUER"), JwtSecret: []byte(os.Getenv("JWT_SECRET"))}

	authUsecase := usecase.NewAuthUsecase(jwtMeta, userRepo, tokenRepo, buyerRepo)
	authHandler := delivery.NewAuthHandler(jwtMeta, userRepo, tokenRepo, buyerRepo, baseLogger, authUsecase)

	mux := http.NewServeMux()

	loginHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Login),
	)
	mux.Handle("/auth/login", loginHandler)

	registerHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Register),
	)
	mux.Handle("/auth/register", registerHandler)

	refreshHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Refresh),
	)
	mux.Handle("/auth/refresh", refreshHandler)

	logoutHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Logout),
	)
	mux.Handle("/auth/logout", logoutHandler)

	resetPasswordHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.ResetPassword),
	)
	mux.Handle("/auth/password", resetPasswordHandler)
	mux.Handle("/auth/reset-password", resetPasswordHandler)

	return mux
}
