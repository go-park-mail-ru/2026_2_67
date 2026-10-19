package main

import (
	"fmt"
	"net/http"

	"vibe_market/backend/internal/auth"
	authRepo "vibe_market/backend/internal/auth/repo/in_memory"
	"vibe_market/backend/internal/buyers"
	buyerRepo "vibe_market/backend/internal/buyers/repo/in_memory"
	"vibe_market/backend/internal/common"

	"vibe_market/backend/internal/products"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		return
	}

	userRepo := authRepo.NewUserInMemoryDB()
	tokenRepo := authRepo.NewRefreshTokenInMemoryDB()
	buyerRepo := buyerRepo.NewBuyerInMemoryDB(userRepo)

	mainMux := http.NewServeMux()

	authMux := auth.NewServeMux(userRepo, tokenRepo, buyerRepo)
	mainMux.Handle("/api/v1/auth/", http.StripPrefix("/api/v1/auth", authMux))

	productsStorage := products.NewInMemoryDB()
	productsHandler := products.MakeProductsHandler(productsStorage)

	productsChain := common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(productsHandler.GetProductsHandler),
	)
	mainMux.Handle("/api/v1/products/", http.StripPrefix("/api/v1/products", productsChain))

	buyersMux := buyers.NewServeMux(buyerRepo)
	mainMux.Handle("/api/v1/buyers/", http.StripPrefix("/api/v1/buyers", buyersMux))

	handler := common.CORSMiddleware(mainMux)

	fmt.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
