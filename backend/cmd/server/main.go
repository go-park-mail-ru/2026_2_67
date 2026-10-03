package main

import (
	"fmt"
	"net/http"

	"myapp/backend/internal/auth"
	"myapp/backend/internal/common"
)

func EchoAccessTokenHandler(w http.ResponseWriter, r *http.Request) {
	accessToken, ok := r.Context().Value(auth.AccessTokenPayloadKey).(*auth.AccessTokenPayload)

	if !ok {
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	fmt.Println("Used AccessTokenPayloadMiddleware")
	fmt.Printf("accessToken: %+v\n", accessToken)

	w.Write([]byte("Check console"))
}

func main() {
	secret := []byte("SECRET")
	issuer := "BmstuOzon"

	// 1. Инициализируем хранилище (подставьте вашу реализацию Storage)
	storage := auth.NewInMemoryDB()

	// 2. Создаем экземпляр AuthHandler
	authHandler := auth.NewAuthHandler(secret, issuer, storage)

	// 3. Создаем роутер (ServeMux)
	mux := http.NewServeMux()

	// EchoAccessTokenHandler
	echoHandler := common.RequireHTTPMethodMiddleware(http.MethodGet)(
		auth.AccessTokenPayloadMiddleware(secret)(http.HandlerFunc(EchoAccessTokenHandler)),
	)
	mux.Handle("/", echoHandler)

	// LoginHandler
	loginHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		auth.AccessTokenPayloadMiddleware(secret)(http.HandlerFunc(authHandler.Login)),
	)
	mux.Handle("/auth/login", loginHandler)

	// RegisterHandler
	registerHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		auth.AccessTokenPayloadMiddleware(secret)(http.HandlerFunc(authHandler.Register)),
	)
	mux.Handle("/auth/register", registerHandler)

	refreshHandler := common.RequireHTTPMethodMiddleware(http.MethodPost)(
		http.HandlerFunc(authHandler.Refresh),
	)
	mux.Handle("/auth/refresh", refreshHandler)

	fmt.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
