package main

import (
	"fmt"
	"net/http"

	"myapp/backend/internal/auth"
	"myapp/backend/internal/common"
)

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	accessToken, ok := r.Context().Value(auth.AccessTokenPayloadKey).(*auth.AccessTokenPayload)

	if !ok {
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	fmt.Println("Has accessToken:", ok)
	fmt.Printf("accessToken: %+v\n", accessToken)

	w.Write([]byte("Check console"))
}

func main() {
	secret := []byte("OKAK")

	// 1. Инициализируем хранилище (подставьте вашу реализацию Storage)
	var storage auth.Storage // e.g. storage := repository.NewStorage(...)

	// 2. Создаем экземпляр AuthHandler
	authHandler := auth.MakeAuthHandler(secret, "ozon", storage)

	// 3. Создаем роутер (ServeMux)
	mux := http.NewServeMux()

	// 4. Регистрируем EchoHandler
	echoChain := common.RequireHTTPMethod(http.MethodGet)(
		auth.AccessTokenMiddleware(secret)(http.HandlerFunc(EchoHandler)),
	)
	mux.Handle("/", echoChain)

	// 5. Регистрируем LoginHandler для пути /auth/login
	// Обратите внимание: для логина обычно используется POST
	loginChain := common.RequireHTTPMethod(http.MethodPost)(
		auth.AccessTokenMiddleware(secret)(http.HandlerFunc(authHandler.LoginHandler)),
	)
	mux.Handle("/auth/login", loginChain)

	fmt.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
