package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	"vibe_market/backend/internal/onlineconf/domain"

	"go.uber.org/zap"
)

type configItemResponse struct {
	ID          int64           `json:"id"`
	Key         string          `json:"key"`
	Scope       string          `json:"scope"`
	ValueType   string          `json:"valueType"`
	Value       json.RawMessage `json:"value"`
	Description string          `json:"description"`
	CreatedBy   int64           `json:"createdBy"`
	UpdatedBy   int64           `json:"updatedBy"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type configsResponse struct {
	Configs []configItemResponse `json:"configs"`
}

type postConfigRequest struct {
	Key         *string          `json:"key"`
	Scope       *string          `json:"scope"`
	ValueType   *string          `json:"valueType"`
	Value       *json.RawMessage `json:"value"`
	Description string           `json:"description"`
}

type putConfigRequest struct {
	ValueType   *string          `json:"valueType"`
	Value       *json.RawMessage `json:"value"`
	Description string           `json:"description"`
}

type ConfigsHandler struct {
	configUsecase domain.ConfigUsecase
	baseLogger    *zap.Logger
}

func NewConfigsHandler(configUsecase domain.ConfigUsecase, logger *zap.Logger) *ConfigsHandler {
	return &ConfigsHandler{
		configUsecase: configUsecase,
		baseLogger:    logger,
	}
}

func toConfigResponse(cfg domain.Config) configItemResponse {
	return configItemResponse{
		ID:          cfg.ID,
		Key:         cfg.Key,
		Scope:       cfg.Scope,
		ValueType:   cfg.ValueType,
		Value:       cfg.Value,
		Description: cfg.Description,
		CreatedBy:   cfg.CreatedBy,
		UpdatedBy:   cfg.UpdatedBy,
		CreatedAt:   cfg.CreatedAt,
		UpdatedAt:   cfg.UpdatedAt,
	}
}

func (h *ConfigsHandler) Configs(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context(), h.baseLogger)
	query := r.URL.Query()

	configs, err := h.configUsecase.List(query.Get("scope"), query.Get("q"))
	if err != nil {
		logger.Error("ошибка получения списка конфигов", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	items := make([]configItemResponse, 0, len(configs))
	for _, cfg := range configs {
		items = append(items, toConfigResponse(cfg))
	}

	writeJSON(w, http.StatusOK, configsResponse{Configs: items})
}

func (h *ConfigsHandler) PostConfig(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context(), h.baseLogger)
	admin, ok := adminFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, domain.ErrAdminForbidden.Error())
		return
	}

	var req postConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный формат тела запроса")
		return
	}

	if req.Key == nil || req.Scope == nil || req.ValueType == nil || req.Value == nil {
		writeError(w, http.StatusUnprocessableEntity, domain.ErrConfigInvalid.Error())
		return
	}

	created, err := h.configUsecase.Create(domain.CreateConfigInDTO{
		Key:         *req.Key,
		Scope:       *req.Scope,
		ValueType:   *req.ValueType,
		Value:       *req.Value,
		Description: req.Description,
		AdminUserID: admin.UserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConfigInvalid), errors.Is(err, domain.ErrConfigValueType),
			errors.Is(err, domain.ErrKeyInvalid), errors.Is(err, domain.ErrScopeInvalid):
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, domain.ErrConfigAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		default:
			logger.Error("ошибка создания конфига", zap.Error(err))
			writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		}
		return
	}

	writeJSON(w, http.StatusCreated, toConfigResponse(created))
}

func (h *ConfigsHandler) PutConfig(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context(), h.baseLogger)
	admin, ok := adminFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, domain.ErrAdminForbidden.Error())
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusNotFound, domain.ErrConfigNotFound.Error())
		return
	}

	var req putConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный формат тела запроса")
		return
	}

	if req.ValueType == nil || req.Value == nil {
		writeError(w, http.StatusUnprocessableEntity, domain.ErrConfigInvalid.Error())
		return
	}

	updated, err := h.configUsecase.Update(domain.UpdateConfigInDTO{
		ID:          id,
		ValueType:   *req.ValueType,
		Value:       *req.Value,
		Description: req.Description,
		AdminUserID: admin.UserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConfigNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrConfigInvalid), errors.Is(err, domain.ErrConfigValueType):
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			logger.Error("ошибка обновления конфига", zap.Int64("id", id), zap.Error(err))
			writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		}
		return
	}

	writeJSON(w, http.StatusOK, toConfigResponse(updated))
}

func (h *ConfigsHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context(), h.baseLogger)
	admin, ok := adminFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, domain.ErrAdminForbidden.Error())
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusNotFound, domain.ErrConfigNotFound.Error())
		return
	}

	if err := h.configUsecase.Delete(id, admin.UserID); err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		logger.Error("ошибка удаления конфига", zap.Int64("id", id), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
