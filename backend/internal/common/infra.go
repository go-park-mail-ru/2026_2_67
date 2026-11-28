package common

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const dbMaxOpenConn = 20

// NewRedisClient создает и проверяет подключение к Redis из переменных окружения.
func NewRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: "",
		DB:       0,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		panic(err)
	}
	return rdb
}

// NewPostgresClient создает и проверяет пул подключений к PostgreSQL из переменных окружения.
func NewPostgresClient() *sql.DB {
	connURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("pgx", connURL)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(dbMaxOpenConn)
	if err = db.Ping(); err != nil {
		panic(err)
	}

	return db
}

// NewLogger создает zap.Logger в режиме production или development в зависимости от LOGGER.
func NewLogger() (*zap.Logger, error) {
	if strings.ToLower(os.Getenv("LOGGER")) == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
