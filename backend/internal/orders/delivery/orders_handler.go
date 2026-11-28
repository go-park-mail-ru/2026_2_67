package delivery

import (
	"context"
	"encoding/json"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/orders/usecase"

	"go.uber.org/zap"
)

type OnlineConfClient interface {
	GetConfigJSON(scope, key string) (json.RawMessage, bool)
	IsFlagEnabled(scope, key string, userID int64) bool
}

type OrdersHandler struct {
	orderRepo  domain.OrderRepo
	usecase    usecase.OrderUsecase
	onlineConf OnlineConfClient
	baseLogger *zap.Logger
}

func NewOrdersHandler(orderRepo domain.OrderRepo, usecase usecase.OrderUsecase, onlineConf OnlineConfClient, baseLogger *zap.Logger) *OrdersHandler {
	return &OrdersHandler{orderRepo: orderRepo, usecase: usecase, onlineConf: onlineConf, baseLogger: baseLogger}
}

func (h *OrdersHandler) logger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(common.LoggerContextKey).(*zap.Logger)
	if !ok {
		return h.baseLogger
	}
	if logger != nil {
		return logger
	}
	return h.baseLogger
}
