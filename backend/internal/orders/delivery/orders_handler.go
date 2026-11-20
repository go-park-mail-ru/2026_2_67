package delivery

import (
	"context"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/orders/domain"

	"go.uber.org/zap"
)

type OrdersHandler struct {
	orderRepo  domain.OrderRepo
	baseLogger *zap.Logger
}

func NewOrdersHandler(
	orderRepo domain.OrderRepo,
	_ []byte,
	baseLogger *zap.Logger,
) *OrdersHandler {
	return &OrdersHandler{
		orderRepo:  orderRepo,
		baseLogger: baseLogger,
	}
}

func (h *OrdersHandler) logger(ctx context.Context) *zap.Logger {
	if val := ctx.Value(common.LoggerContextKey); val != nil {
		if l, ok := val.(*zap.Logger); ok && l != nil {
			return l
		}
	}
	return h.baseLogger
}
