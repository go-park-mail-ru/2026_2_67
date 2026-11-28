package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"vibe_market/backend/internal/auth"
	authRepoPostgres "vibe_market/backend/internal/auth/repository/postgres"
	authRepoRedis "vibe_market/backend/internal/auth/repository/redis"
	"vibe_market/backend/internal/buyers"
	buyerRepoStorage "vibe_market/backend/internal/buyers/repository/minio"
	buyerRepoPostgres "vibe_market/backend/internal/buyers/repository/postgres"
	"vibe_market/backend/internal/categories"
	categoryRepoPostgres "vibe_market/backend/internal/categories/repository/postgres"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/onlineconf_client"
	"vibe_market/backend/internal/orders"
	orderRepoPostgres "vibe_market/backend/internal/orders/repository/postgres"
	"vibe_market/backend/internal/pickup_points"
	pickupPointRepoPostgres "vibe_market/backend/internal/pickup_points/repository/postgres"
	"vibe_market/backend/internal/products"
	productRepoPostgres "vibe_market/backend/internal/products/repository/postgres"
	"vibe_market/backend/internal/sellers"
	sellerRepoPostgres "vibe_market/backend/internal/sellers/repository/postgres"

	"github.com/joho/godotenv"
	minioClient "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"go.uber.org/zap"
)

func main() {
	logger, err := common.NewLogger()
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

	redisClient := common.NewRedisClient()
	logger.Info("Redis активен")

	postgresClient := common.NewPostgresClient()
	logger.Info("Postgres активен")

	userRepo := authRepoPostgres.NewUserRepo(postgresClient)
	tokenRepo := authRepoRedis.NewRefreshTokenRepo(redisClient)
	buyerRepo := buyerRepoPostgres.NewBuyerRepo(postgresClient)
	basketRepo := buyerRepoPostgres.NewBasketRepo(postgresClient)
	orderRepo := orderRepoPostgres.NewOrderRepo(postgresClient)
	pickupPointRepo := pickupPointRepoPostgres.NewPickupPointRepo(postgresClient)
	productRepo := productRepoPostgres.NewProductRepo(postgresClient)
	productReviewRepo := productRepoPostgres.NewProductReviewRepo(postgresClient)
	sellerRepo := sellerRepoPostgres.NewSellerRepo(postgresClient)
	categoryRepo := categoryRepoPostgres.NewCategoryRepo(postgresClient)
	minio, err := minioClient.New(fmt.Sprintf("%s:%s", os.Getenv("STORAGE_HOST"), os.Getenv("STORAGE_PORT")), &minioClient.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("STORAGE_USER"), os.Getenv("STORAGE_PASSWORD"), ""),
		Secure: false,
	})
	if err != nil {
		logger.Fatal("Не удалось создать S3-клиент", zap.Error(err))
	}
	avatarStorage := buyerRepoStorage.NewBuyerAvatarRepo(minio, os.Getenv("STORAGE_BUCKET"))

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	ocClient := onlineconf_client.NewClient(os.Getenv("ONLINECONF_URL"), logger)
	ocClient.Start(context.Background())

	mainMux := http.NewServeMux()

	authMux := auth.NewServeMux(userRepo, tokenRepo, buyerRepo, logger)
	mainMux.Handle("/api/v1/auth/", http.StripPrefix("/api/v1/auth", authMux))

	productsMux := products.NewServeMux(productRepo, productReviewRepo, orderRepo, ocClient, logger, jwtSecret)
	mainMux.Handle("/api/v1/products/", http.StripPrefix("/api/v1/products", productsMux))

	categoriesMux := categories.NewServeMux(categoryRepo, logger)
	mainMux.Handle("/api/v1/categories/", http.StripPrefix("/api/v1/categories", categoriesMux))

	buyersMux := buyers.NewServeMux(buyerRepo, basketRepo, userRepo, jwtSecret, logger, avatarStorage)
	mainMux.Handle("/api/v1/buyers/", http.StripPrefix("/api/v1/buyers", buyersMux))

	ordersMux := orders.NewServeMux(orderRepo, jwtSecret, ocClient, logger)
	mainMux.Handle("/api/v1/orders/", http.StripPrefix("/api/v1/orders", ordersMux))

	pickupPointsMux := pickup_points.NewServeMux(pickupPointRepo, logger)
	mainMux.Handle("/api/v1/pickup_points/", http.StripPrefix("/api/v1/pickup_points", pickupPointsMux))

	sellersMux := sellers.NewServeMux(sellerRepo, logger)
	mainMux.Handle("/api/v1/sellers/", http.StripPrefix("/api/v1/sellers", sellersMux))

	handler := common.EndpointLogMiddleware(logger)(common.CORSMiddleware(mainMux))

	port, err := strconv.ParseUint(os.Getenv("APP_PORT"), 10, 32)
	if err != nil {
		logger.Error("Неправильный порт", zap.String("port", os.Getenv("APP_PORT")))
		os.Exit(1)
	}

	logger.Info("Сервер запущен", zap.Uint64("port", port))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		logger.Fatal("Ошибка сервера", zap.Error(err))
	}
}
