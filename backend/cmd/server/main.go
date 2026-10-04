package main

import (
	"fmt"
	"net/http"
	"os"

	"bmstuozon/backend/internal/auth"
	"bmstuozon/backend/internal/common"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		return
	}
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	jwtIssuer := os.Getenv("JWT_ISSUER")

	storage := auth.NewInMemoryDB()

	authHandler := auth.NewAuthHandler(jwtSecret, jwtIssuer, storage)

	mux := http.NewServeMux()

	// Login
	loginHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Login),
	)
	mux.Handle("/api/v1/auth/login", loginHandler)

	// Register
	registerHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Register),
	)
	mux.Handle("/api/v1/auth/register", registerHandler)

	// Refresh
	refreshHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Refresh),
	)
	mux.Handle("/api/v1/auth/refresh", refreshHandler)

	// Logout
	logoutHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Logout),
	)
	mux.Handle("/api/v1/auth/logout", logoutHandler)

	fmt.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
