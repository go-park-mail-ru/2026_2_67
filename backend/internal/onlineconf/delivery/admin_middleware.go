package delivery

import (
	"context"
	"net/http"
	authDomain "vibe_market/backend/internal/auth/domain"
	authMiddleware "vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/onlineconf/domain"

	"go.uber.org/zap"
)

type meResponse struct {
	UserID int64  `json:"userId"`
	Role   string `json:"role"`
}

type AdminsHandler struct {
	adminUsecase domain.AdminUsecase
	baseLogger   *zap.Logger
}

func NewAdminsHandler(adminUsecase domain.AdminUsecase, logger *zap.Logger) *AdminsHandler {
	return &AdminsHandler{
		adminUsecase: adminUsecase,
		baseLogger:   logger,
	}
}

func AdminMiddleware(jwtSecret []byte, adminUsecase domain.AdminUsecase, baseLogger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		authHandler := authMiddleware.DefaultAuthMiddleware(jwtSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := loggerFromContext(r.Context(), baseLogger)

			payload, ok := r.Context().Value(authMiddleware.AccessTokenPayloadKey).(*authDomain.AccessTokenPayload)
			if !ok || payload == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			admin, err := adminUsecase.GetByUserID(payload.UserID)
			if err != nil {
				logger.Warn("доступ к админке onlineconf запрещён", zap.Int64("user_id", payload.UserID), zap.Error(err))
				writeError(w, http.StatusForbidden, domain.ErrAdminForbidden.Error())
				return
			}

			ctx := context.WithValue(r.Context(), AdminContextKey, admin)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))

		return authHandler
	}
}

func (h *AdminsHandler) Me(w http.ResponseWriter, r *http.Request) {
	admin, ok := adminFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, domain.ErrAdminForbidden.Error())
		return
	}

	writeJSON(w, http.StatusOK, meResponse{
		UserID: admin.UserID,
		Role:   admin.Role,
	})
}
