package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"vibe_market/backend/internal/auth"
	authRepo "vibe_market/backend/internal/auth/repo/in_memory"
	"vibe_market/backend/internal/buyers"
	buyerRepo "vibe_market/backend/internal/buyers/repo/in_memory"
	"vibe_market/backend/internal/common"

	"vibe_market/backend/internal/products"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: "",
		DB:       0})

	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Redis подключен:", pong)
	return rdb
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Загружен .env")

	userRepo := authRepo.NewUserInMemoryDB()
	tokenRepo := authRepo.NewRefreshTokenInMemoryDB()
	buyerRepo := buyerRepo.NewBuyerInMemoryDB(userRepo)

	mainMux := http.NewServeMux()

	authMux := auth.NewServeMux(userRepo, tokenRepo)
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

	port, err := strconv.ParseInt(os.Getenv("APP_PORT"), 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Сервер запущен на :%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		fmt.Printf("Ошибка сервера: %v\n", err)
	}
}
