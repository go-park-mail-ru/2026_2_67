package onlineconf

import (
	"context"
	"net/http"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/onlineconf/delivery"
	"vibe_market/backend/internal/onlineconf/domain"
	"vibe_market/backend/internal/onlineconf/usecase"

	"go.uber.org/zap"
)

func NewServeMux(
	ctx context.Context,
	adminRepo domain.AdminRepo,
	configRepo domain.ConfigRepo,
	flagRepo domain.FeatureFlagRepo,
	historyRepo domain.HistoryRepo,
	logger *zap.Logger,
	jwtSecret []byte,
) *http.ServeMux {
	usecases := usecase.NewUsecases(adminRepo, configRepo, flagRepo, historyRepo, logger)
	usecases.Cache.Start(ctx)

	adminGuard := delivery.AdminMiddleware(jwtSecret, usecases.Admin, logger)

	adminsHandler := delivery.NewAdminsHandler(usecases.Admin, logger)
	configsHandler := delivery.NewConfigsHandler(usecases.Config, logger)
	flagsHandler := delivery.NewFlagsHandler(usecases.FeatureFlag, logger)
	historyHandler := delivery.NewHistoryHandler(usecases.History, logger)
	runtimeHandler := delivery.NewRuntimeHandler(usecases.Runtime, jwtSecret, logger)

	mux := http.NewServeMux()

	mux.Handle("/api/v1/onlineconf/admin/me", common.RequireHTTPMethodMiddleware(http.MethodGet)(
		adminGuard(http.HandlerFunc(adminsHandler.Me)),
	))

	mux.Handle("/api/v1/onlineconf/admin/configs", adminGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			configsHandler.Configs(w, r)
		case http.MethodPost:
			configsHandler.PostConfig(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/onlineconf/admin/configs/{id}", adminGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			configsHandler.PutConfig(w, r)
		case http.MethodDelete:
			configsHandler.DeleteConfig(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/onlineconf/admin/flags", adminGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			flagsHandler.Flags(w, r)
		case http.MethodPost:
			flagsHandler.PostFlag(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/onlineconf/admin/flags/{id}", adminGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			flagsHandler.PutFlag(w, r)
		case http.MethodDelete:
			flagsHandler.DeleteFlag(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/onlineconf/admin/flags/{id}/evaluate", common.RequireHTTPMethodMiddleware(http.MethodGet)(
		adminGuard(http.HandlerFunc(flagsHandler.EvaluateFlag)),
	))

	mux.Handle("/api/v1/onlineconf/admin/history", common.RequireHTTPMethodMiddleware(http.MethodGet)(
		adminGuard(http.HandlerFunc(historyHandler.History)),
	))

	mux.Handle("/api/v1/onlineconf/runtime/frontend", common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(runtimeHandler.FrontendRuntime),
	))

	mux.Handle("/api/v1/onlineconf/snapshot", common.RequireHTTPMethodMiddleware(http.MethodGet)(
		http.HandlerFunc(runtimeHandler.Snapshot),
	))

	return mux
}
