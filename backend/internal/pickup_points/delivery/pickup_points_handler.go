package delivery

import (
	"context"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/pickup_points/domain"

	"go.uber.org/zap"
)

type PickupPointsHandler struct {
	pickupPointRepo domain.PickupPointRepo
	baseLogger      *zap.Logger
}

func NewPickupPointsHandler(repo domain.PickupPointRepo, logger *zap.Logger) *PickupPointsHandler {
	return &PickupPointsHandler{pickupPointRepo: repo, baseLogger: logger}
}

func (h *PickupPointsHandler) logger(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(common.LoggerContextKey).(*zap.Logger); ok && logger != nil {
		return logger
	}
	return h.baseLogger
}
