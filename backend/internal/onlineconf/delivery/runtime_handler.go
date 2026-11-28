package delivery

import (
	"encoding/json"
	"hash/fnv"
	"net/http"
	"strconv"
	authMiddleware "vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/onlineconf/domain"

	"go.uber.org/zap"
)

type runtimeFrontendResponse struct {
	Configs map[string]json.RawMessage `json:"configs"`
	Flags   map[string]bool            `json:"flags"`
}

type snapshotConfigResponse struct {
	Scope     string          `json:"scope"`
	Key       string          `json:"key"`
	ValueType string          `json:"valueType"`
	Value     json.RawMessage `json:"value"`
}

type snapshotFlagResponse struct {
	Scope          string  `json:"scope"`
	Key            string  `json:"key"`
	Enabled        bool    `json:"enabled"`
	RolloutPercent int     `json:"rolloutPercent"`
	Whitelist      []int64 `json:"whitelist"`
	Blacklist      []int64 `json:"blacklist"`
}

type snapshotResponse struct {
	Configs []snapshotConfigResponse `json:"configs"`
	Flags   []snapshotFlagResponse   `json:"flags"`
}

type RuntimeHandler struct {
	runtimeUsecase domain.RuntimeUsecase
	jwtSecret      []byte
	baseLogger     *zap.Logger
}

func NewRuntimeHandler(runtimeUsecase domain.RuntimeUsecase, jwtSecret []byte, logger *zap.Logger) *RuntimeHandler {
	return &RuntimeHandler{
		runtimeUsecase: runtimeUsecase,
		jwtSecret:      jwtSecret,
		baseLogger:     logger,
	}
}

func (h *RuntimeHandler) resolveUserID(r *http.Request) int64 {
	if payload, err := authMiddleware.ParseAccessTokenPayload(r, h.jwtSecret); err == nil && payload != nil {
		return payload.UserID
	}

	if deviceID := r.Header.Get("X-Device-Id"); deviceID != "" {
		if parsed, err := strconv.ParseInt(deviceID, 10, 64); err == nil && parsed >= 0 {
			return parsed
		}
		hash := fnv.New32a()
		_, _ = hash.Write([]byte(deviceID))
		return int64(hash.Sum32())
	}

	return 0
}

func (h *RuntimeHandler) FrontendRuntime(w http.ResponseWriter, r *http.Request) {
	userID := h.resolveUserID(r)
	res := h.runtimeUsecase.FrontendRuntime(userID)

	writeJSON(w, http.StatusOK, runtimeFrontendResponse{
		Configs: res.Configs,
		Flags:   res.Flags,
	})
}

func (h *RuntimeHandler) Snapshot(w http.ResponseWriter, r *http.Request) {
	snapshot := h.runtimeUsecase.Snapshot()

	configs := make([]snapshotConfigResponse, 0, len(snapshot.Configs))
	for _, cfg := range snapshot.Configs {
		configs = append(configs, snapshotConfigResponse{
			Scope:     cfg.Scope,
			Key:       cfg.Key,
			ValueType: cfg.ValueType,
			Value:     cfg.Value,
		})
	}

	flags := make([]snapshotFlagResponse, 0, len(snapshot.Flags))
	for _, flag := range snapshot.Flags {
		flags = append(flags, snapshotFlagResponse{
			Scope:          flag.Scope,
			Key:            flag.Key,
			Enabled:        flag.Enabled,
			RolloutPercent: flag.RolloutPercent,
			Whitelist:      flag.Whitelist,
			Blacklist:      flag.Blacklist,
		})
	}

	writeJSON(w, http.StatusOK, snapshotResponse{
		Configs: configs,
		Flags:   flags,
	})
}
