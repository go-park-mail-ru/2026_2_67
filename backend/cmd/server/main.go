package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"vibe_market/backend/internal/auth"
	authRepoPostgres "vibe_market/backend/internal/auth/repository/postgres"
	authRepoRedis "vibe_market/backend/internal/auth/repository/redis"
	"vibe_market/backend/internal/buyers"
	buyerRepoPostgres "vibe_market/backend/internal/buyers/repository/postgres"
	"vibe_market/backend/internal/common"
	productRepoPostgres "vibe_market/backend/internal/products/repository/postgres"

	"vibe_market/backend/internal/products"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"go.uber.org/zap"
)

const dbMaxOpenConn = 20

func NewRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: "",
		DB:       0})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	return rdb
}

func NewPostgresClient() *sql.DB {
	connURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

	db, err := sql.Open("pgx", connURL)

	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(dbMaxOpenConn)
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

func NewLogger() (*zap.Logger, error) {
	if strings.ToLower(os.Getenv("LOGGER")) == "production" {
		return zap.NewProduction()
	} else {
		return zap.NewDevelopment()
	}
}

func main() {
	logger, err := NewLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	err = godotenv.Load()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info(".env загружен")
	logger.Info("logger создан")

	redisClient := NewRedisClient()
	logger.Info("Redis активен")

	postgresClient := NewPostgresClient()
	logger.Info("Postgres активен")

	userRepo := authRepoPostgres.NewUserRepo(postgresClient)
	tokenRepo := authRepoRedis.NewRefreshTokenRepo(redisClient)
	buyerRepo := buyerRepoPostgres.NewBuyerRepo(postgresClient)
	productsRepo := productRepoPostgres.NewProductRepo(postgresClient)

	mainMux := http.NewServeMux()

	authMux := auth.NewServeMux(userRepo, tokenRepo, buyerRepo, logger)
	mainMux.Handle("/api/v1/auth/", http.StripPrefix("/api/v1/auth", authMux))

	productsMux := products.NewServeMux(productsRepo, logger)
	mainMux.Handle("/api/v1/products/", http.StripPrefix("/api/v1/products", productsMux))

	buyersMux := buyers.NewServeMux(buyerRepo, logger)
	mainMux.Handle("/api/v1/buyers/", http.StripPrefix("/api/v1/buyers", buyersMux))

	handler := common.EndpointLogMiddleware(logger)(common.CORSMiddleware(mainMux))

	port, err := strconv.ParseInt(os.Getenv("APP_PORT"), 10, 64)
	if err != nil {
		logger.Error("Wrong port", zap.String("port", os.Getenv("APP_PORT")))
		os.Exit(1)
	}

	logger.Info("Сервер запущен", zap.Int64("port", port))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		logger.Fatal("Ошибка сервера", zap.Error(err))
	}
}
