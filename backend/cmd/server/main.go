package main

import (
	"fmt"
	"net/http"

	"myapp/backend/internal/auth"
	"myapp/backend/internal/common"
	"myapp/backend/internal/products"
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
	authHandler := auth.MakeAuthHandler(secret, issuer, storage)

	// 3. Создаем роутер (ServeMux)
	mux := http.NewServeMux()

	// 4. Регистрируем EchoAccessTokenHandler
	echoChain := common.RequireHTTPMethod(http.MethodGet)(
		auth.AccessTokenPayloadMiddleware(secret)(http.HandlerFunc(EchoAccessTokenHandler)),
	)
	mux.Handle("/", echoChain)

	// 5. Регистрируем LoginHandler для пути /auth/login
	// Обратите внимание: для логина обычно используется POST
	loginChain := common.RequireHTTPMethod(http.MethodPost)(
		auth.AccessTokenPayloadMiddleware(secret)(http.HandlerFunc(authHandler.LoginHandler)),
	)
	mux.Handle("/auth/login", loginChain)

	// 6. Регистрируем RegisterHandler для /auth/register
	registerChain := common.RequireHTTPMethod(http.MethodPost)(
		auth.AccessTokenPayloadMiddleware(secret)(http.HandlerFunc(authHandler.RegisterHandler)),
	)
	mux.Handle("/auth/register", registerChain)

	// 7. Инициализируем хранилище и обработчик товаров
	productsStorage := products.NewInMemoryDB()
	productsHandler := products.MakeProductsHandler(productsStorage)

	// 8. Регистрируем ProductsHandler для /api/v1/products
	productsChain := common.RequireHTTPMethod(http.MethodGet)(
		auth.AccessTokenPayloadMiddleware(secret)(http.HandlerFunc(productsHandler.GetProductsHandler)),
	)
	mux.Handle("/api/v1/products", productsChain)

	fmt.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
