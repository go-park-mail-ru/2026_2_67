package delivery

import (
	"context"
	"encoding/json"
	"net/http"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/onlineconf/domain"

	"go.uber.org/zap"
)

type contextKey string

const AdminContextKey contextKey = "OnlineConfAdmin"

type errorResponse struct {
	ErrMessage string `json:"errMessage"`
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, errorResponse{ErrMessage: message})
}

func loggerFromContext(ctx context.Context, fallback *zap.Logger) *zap.Logger {
	if logger, ok := ctx.Value(common.LoggerContextKey).(*zap.Logger); ok && logger != nil {
		return logger
	}
	return fallback
}

func adminFromContext(ctx context.Context) (domain.Admin, bool) {
	admin, ok := ctx.Value(AdminContextKey).(domain.Admin)
	return admin, ok
}
