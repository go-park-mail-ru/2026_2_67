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

type flagItemResponse struct {
	ID             int64     `json:"id"`
	Key            string    `json:"key"`
	Scope          string    `json:"scope"`
	Enabled        bool      `json:"enabled"`
	RolloutPercent int       `json:"rolloutPercent"`
	Whitelist      []int64   `json:"whitelist"`
	Blacklist      []int64   `json:"blacklist"`
	Description    string    `json:"description"`
	CreatedBy      int64     `json:"createdBy"`
	UpdatedBy      int64     `json:"updatedBy"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type flagsResponse struct {
	Flags []flagItemResponse `json:"flags"`
}

type postFlagRequest struct {
	Key            *string `json:"key"`
	Scope          *string `json:"scope"`
	Enabled        *bool   `json:"enabled"`
	RolloutPercent *int    `json:"rolloutPercent"`
	Whitelist      []int64 `json:"whitelist"`
	Blacklist      []int64 `json:"blacklist"`
	Description    string  `json:"description"`
}

type putFlagRequest struct {
	Enabled        *bool   `json:"enabled"`
	RolloutPercent *int    `json:"rolloutPercent"`
	Whitelist      []int64 `json:"whitelist"`
	Blacklist      []int64 `json:"blacklist"`
	Description    string  `json:"description"`
}

type evaluateFlagResponse struct {
	FlagID  int64  `json:"flagId"`
	Key     string `json:"key"`
	Scope   string `json:"scope"`
	UserID  int64  `json:"userId"`
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
	Bucket  int    `json:"bucket"`
}

type FlagsHandler struct {
	flagUsecase domain.FeatureFlagUsecase
	baseLogger  *zap.Logger
}

func NewFlagsHandler(flagUsecase domain.FeatureFlagUsecase, logger *zap.Logger) *FlagsHandler {
	return &FlagsHandler{
		flagUsecase: flagUsecase,
		baseLogger:  logger,
	}
}

func toFlagResponse(flag domain.FeatureFlag) flagItemResponse {
	whitelist := flag.Whitelist
	if whitelist == nil {
		whitelist = []int64{}
	}
	blacklist := flag.Blacklist
	if blacklist == nil {
		blacklist = []int64{}
	}
	return flagItemResponse{
		ID:             flag.ID,
		Key:            flag.Key,
		Scope:          flag.Scope,
		Enabled:        flag.Enabled,
		RolloutPercent: flag.RolloutPercent,
		Whitelist:      whitelist,
		Blacklist:      blacklist,
		Description:    flag.Description,
		CreatedBy:      flag.CreatedBy,
		UpdatedBy:      flag.UpdatedBy,
		CreatedAt:      flag.CreatedAt,
		UpdatedAt:      flag.UpdatedAt,
	}
}

func (h *FlagsHandler) Flags(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context(), h.baseLogger)
	query := r.URL.Query()

	flags, err := h.flagUsecase.List(query.Get("scope"), query.Get("q"))
	if err != nil {
		logger.Error("ошибка получения списка фича-флагов", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	items := make([]flagItemResponse, 0, len(flags))
	for _, flag := range flags {
		items = append(items, toFlagResponse(flag))
	}

	writeJSON(w, http.StatusOK, flagsResponse{Flags: items})
}

func (h *FlagsHandler) PostFlag(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context(), h.baseLogger)
	admin, ok := adminFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, domain.ErrAdminForbidden.Error())
		return
	}

	var req postFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный формат тела запроса")
		return
	}

	if req.Key == nil || req.Scope == nil || req.Enabled == nil || req.RolloutPercent == nil {
		writeError(w, http.StatusUnprocessableEntity, domain.ErrFeatureFlagInvalid.Error())
		return
	}

	created, err := h.flagUsecase.Create(domain.CreateFeatureFlagInDTO{
		Key:            *req.Key,
		Scope:          *req.Scope,
		Enabled:        *req.Enabled,
		RolloutPercent: *req.RolloutPercent,
		Whitelist:      req.Whitelist,
		Blacklist:      req.Blacklist,
		Description:    req.Description,
		AdminUserID:    admin.UserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrFeatureFlagInvalid), errors.Is(err, domain.ErrFeatureFlagRolloutInvalid),
			errors.Is(err, domain.ErrKeyInvalid), errors.Is(err, domain.ErrScopeInvalid):
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, domain.ErrFeatureFlagAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		default:
			logger.Error("ошибка создания фича-флага", zap.Error(err))
			writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		}
		return
	}

	writeJSON(w, http.StatusCreated, toFlagResponse(created))
}

func (h *FlagsHandler) PutFlag(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context(), h.baseLogger)
	admin, ok := adminFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, domain.ErrAdminForbidden.Error())
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusNotFound, domain.ErrFeatureFlagNotFound.Error())
		return
	}

	var req putFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный формат тела запроса")
		return
	}

	if req.Enabled == nil || req.RolloutPercent == nil {
		writeError(w, http.StatusUnprocessableEntity, domain.ErrFeatureFlagInvalid.Error())
		return
	}

	updated, err := h.flagUsecase.Update(domain.UpdateFeatureFlagInDTO{
		ID:             id,
		Enabled:        *req.Enabled,
		RolloutPercent: *req.RolloutPercent,
		Whitelist:      req.Whitelist,
		Blacklist:      req.Blacklist,
		Description:    req.Description,
		AdminUserID:    admin.UserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrFeatureFlagNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrFeatureFlagInvalid), errors.Is(err, domain.ErrFeatureFlagRolloutInvalid):
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			logger.Error("ошибка обновления фича-флага", zap.Int64("id", id), zap.Error(err))
			writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		}
		return
	}

	writeJSON(w, http.StatusOK, toFlagResponse(updated))
}

func (h *FlagsHandler) DeleteFlag(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context(), h.baseLogger)
	admin, ok := adminFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, domain.ErrAdminForbidden.Error())
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusNotFound, domain.ErrFeatureFlagNotFound.Error())
		return
	}

	if err := h.flagUsecase.Delete(id, admin.UserID); err != nil {
		if errors.Is(err, domain.ErrFeatureFlagNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		logger.Error("ошибка удаления фича-флага", zap.Int64("id", id), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FlagsHandler) EvaluateFlag(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context(), h.baseLogger)

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusNotFound, domain.ErrFeatureFlagNotFound.Error())
		return
	}

	rawUserID := r.URL.Query().Get("userId")
	if rawUserID == "" {
		writeError(w, http.StatusBadRequest, "параметр userId обязателен")
		return
	}

	userID, err := strconv.ParseInt(rawUserID, 10, 64)
	if err != nil || userID < 0 {
		writeError(w, http.StatusBadRequest, "некорректный параметр userId")
		return
	}

	res, err := h.flagUsecase.Evaluate(id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrFeatureFlagNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		logger.Error("ошибка вычисления фича-флага", zap.Int64("id", id), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	writeJSON(w, http.StatusOK, evaluateFlagResponse{
		FlagID:  res.FlagID,
		Key:     res.Key,
		Scope:   res.Scope,
		UserID:  res.UserID,
		Enabled: res.Enabled,
		Reason:  res.Reason,
		Bucket:  res.Bucket,
	})
}
