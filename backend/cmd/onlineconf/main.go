package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/onlineconf"
	onlineconfRepoPostgres "vibe_market/backend/internal/onlineconf/repository/postgres"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const (
	defaultOnlineConfPort = uint64(8081)
	shutdownTimeout       = 5 * time.Second
)

func resolvePort() uint64 {
	rawPort := strings.TrimSpace(os.Getenv("ONLINECONF_PORT"))
	if rawPort == "" {
		return defaultOnlineConfPort
	}
	port, err := strconv.ParseUint(rawPort, 10, 32)
	if err != nil || port == 0 {
		return defaultOnlineConfPort
	}
	return port
}

func main() {
	logger, err := common.NewLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Info(".env не загружен, используются переменные окружения", zap.Error(err))
	}

	postgresClient := common.NewPostgresClient()
	defer postgresClient.Close()
	logger.Info("Postgres активен")

	adminRepo := onlineconfRepoPostgres.NewAdminRepo(postgresClient)
	configRepo := onlineconfRepoPostgres.NewConfigRepo(postgresClient)
	flagRepo := onlineconfRepoPostgres.NewFeatureFlagRepo(postgresClient)
	historyRepo := onlineconfRepoPostgres.NewHistoryRepo(postgresClient)

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mux := onlineconf.NewServeMux(ctx, adminRepo, configRepo, flagRepo, historyRepo, logger, jwtSecret)
	handler := common.EndpointLogMiddleware(logger)(common.CORSMiddleware(mux))

	port := resolvePort()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	logger.Info("OnlineConf сервер запущен", zap.Uint64("port", port))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("Ошибка OnlineConf сервера", zap.Error(err))
	}
}
