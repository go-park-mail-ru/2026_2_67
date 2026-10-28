package delivery

import (
	"context"
	"net/http"
	authDomain "vibe_market/backend/internal/auth/domain"
	authMiddleware "vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/orders/domain"

	"go.uber.org/zap"
)

type OrdersHandler struct {
	orderRepo  domain.OrderRepo
	jwtSecret  []byte
	baseLogger *zap.Logger
}

func NewOrdersHandler(
	orderRepo domain.OrderRepo,
	jwtSecret []byte,
	baseLogger *zap.Logger,
) *OrdersHandler {
	return &OrdersHandler{
		orderRepo:  orderRepo,
		jwtSecret:  jwtSecret,
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

func (h *OrdersHandler) authenticate(r *http.Request) (*authDomain.AccessTokenPayload, int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, http.StatusBadRequest, authDomain.ErrUserIDNotFound
	}

	payload, err := authMiddleware.ParseAccessTokenPayload(r, h.jwtSecret)
	if err != nil || payload == nil {
		return nil, http.StatusForbidden, err
	}

	return payload, http.StatusOK, nil
}
