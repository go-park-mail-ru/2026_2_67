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

type historyItemResponse struct {
	ID         int64            `json:"id"`
	EntityType string           `json:"entityType"`
	EntityID   int64            `json:"entityId"`
	EntityKey  string           `json:"entityKey"`
	Scope      string           `json:"scope"`
	Action     string           `json:"action"`
	OldValue   *json.RawMessage `json:"oldValue"`
	NewValue   *json.RawMessage `json:"newValue"`
	ChangedBy  int64            `json:"changedBy"`
	CreatedAt  time.Time        `json:"createdAt"`
}

type historyResponse struct {
	History []historyItemResponse `json:"history"`
}

type HistoryHandler struct {
	historyUsecase domain.HistoryUsecase
	baseLogger     *zap.Logger
}

func NewHistoryHandler(historyUsecase domain.HistoryUsecase, logger *zap.Logger) *HistoryHandler {
	return &HistoryHandler{
		historyUsecase: historyUsecase,
		baseLogger:     logger,
	}
}

func rawPtr(raw json.RawMessage) *json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	cp := make(json.RawMessage, len(raw))
	copy(cp, raw)
	return &cp
}

func (h *HistoryHandler) History(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context(), h.baseLogger)
	query := r.URL.Query()

	var entityID int64
	if rawEntityID := query.Get("entityId"); rawEntityID != "" {
		parsed, err := strconv.ParseInt(rawEntityID, 10, 64)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "некорректный параметр entityId")
			return
		}
		entityID = parsed
	}

	var limit int
	if rawLimit := query.Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, domain.ErrHistoryLimitInvalid.Error())
			return
		}
		limit = parsed
	}

	entries, err := h.historyUsecase.List(domain.ListHistoryInDTO{
		EntityType: query.Get("entityType"),
		EntityID:   entityID,
		Limit:      limit,
	})
	if err != nil {
		if errors.Is(err, domain.ErrHistoryEntityTypeInvalid) || errors.Is(err, domain.ErrHistoryLimitInvalid) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		logger.Error("ошибка получения истории изменений", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	items := make([]historyItemResponse, 0, len(entries))
	for _, entry := range entries {
		items = append(items, historyItemResponse{
			ID:         entry.ID,
			EntityType: entry.EntityType,
			EntityID:   entry.EntityID,
			EntityKey:  entry.EntityKey,
			Scope:      entry.Scope,
			Action:     entry.Action,
			OldValue:   rawPtr(entry.OldValue),
			NewValue:   rawPtr(entry.NewValue),
			ChangedBy:  entry.ChangedBy,
			CreatedAt:  entry.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, historyResponse{History: items})
}
