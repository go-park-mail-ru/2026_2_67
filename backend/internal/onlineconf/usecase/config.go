package usecase

import (
	"bytes"
	"encoding/json"
	"strings"
	"vibe_market/backend/internal/onlineconf/domain"
)

type ConfigUsecase struct {
	configRepo  domain.ConfigRepo
	historyRepo domain.HistoryRepo
	cache       *RuntimeCache
}

func NewConfigUsecase(configRepo domain.ConfigRepo, historyRepo domain.HistoryRepo, cache *RuntimeCache) *ConfigUsecase {
	return &ConfigUsecase{
		configRepo:  configRepo,
		historyRepo: historyRepo,
		cache:       cache,
	}
}

func (u *ConfigUsecase) List(scope string, query string) ([]domain.Config, error) {
	return u.configRepo.List(strings.TrimSpace(scope), strings.TrimSpace(query))
}

func (u *ConfigUsecase) Create(dto domain.CreateConfigInDTO) (domain.Config, error) {
	key := strings.TrimSpace(dto.Key)
	scope := strings.TrimSpace(dto.Scope)
	if !domain.ValidateIdentifier(key) {
		return domain.Config{}, domain.ErrKeyInvalid
	}
	if !domain.ValidateIdentifier(scope) {
		return domain.Config{}, domain.ErrScopeInvalid
	}

	if err := validateConfigValue(dto.ValueType, dto.Value); err != nil {
		return domain.Config{}, err
	}

	created, err := u.configRepo.Create(domain.Config{
		Key:         key,
		Scope:       scope,
		ValueType:   dto.ValueType,
		Value:       bytes.TrimSpace(dto.Value),
		Description: dto.Description,
		CreatedBy:   dto.AdminUserID,
		UpdatedBy:   dto.AdminUserID,
	})
	if err != nil {
		return domain.Config{}, err
	}

	if err := u.recordHistory(domain.HistoryActionCreate, created.ID, created.Key, created.Scope, nil, &created, dto.AdminUserID); err != nil {
		return domain.Config{}, err
	}

	if u.cache != nil {
		u.cache.Invalidate()
	}

	return created, nil
}

func (u *ConfigUsecase) Update(dto domain.UpdateConfigInDTO) (domain.Config, error) {
	if dto.ID <= 0 {
		return domain.Config{}, domain.ErrConfigInvalid
	}

	if err := validateConfigValue(dto.ValueType, dto.Value); err != nil {
		return domain.Config{}, err
	}

	oldConfig, err := u.configRepo.GetByID(dto.ID)
	if err != nil {
		return domain.Config{}, err
	}

	updated, err := u.configRepo.Update(domain.Config{
		ID:          dto.ID,
		ValueType:   dto.ValueType,
		Value:       bytes.TrimSpace(dto.Value),
		Description: dto.Description,
		UpdatedBy:   dto.AdminUserID,
	})
	if err != nil {
		return domain.Config{}, err
	}

	if err := u.recordHistory(domain.HistoryActionUpdate, updated.ID, updated.Key, updated.Scope, &oldConfig, &updated, dto.AdminUserID); err != nil {
		return domain.Config{}, err
	}

	if u.cache != nil {
		u.cache.Invalidate()
	}

	return updated, nil
}

func (u *ConfigUsecase) Delete(id int64, adminUserID int64) error {
	if id <= 0 {
		return domain.ErrConfigNotFound
	}

	deleted, err := u.configRepo.Delete(id)
	if err != nil {
		return err
	}

	if err := u.recordHistory(domain.HistoryActionDelete, deleted.ID, deleted.Key, deleted.Scope, &deleted, nil, adminUserID); err != nil {
		return err
	}

	if u.cache != nil {
		u.cache.Invalidate()
	}

	return nil
}

func (u *ConfigUsecase) recordHistory(
	action string,
	entityID int64,
	entityKey string,
	scope string,
	oldCfg *domain.Config,
	newCfg *domain.Config,
	adminUserID int64,
) error {
	var oldRaw json.RawMessage
	if oldCfg != nil {
		b, err := marshalConfigSnapshot(*oldCfg)
		if err != nil {
			return err
		}
		oldRaw = b
	}

	var newRaw json.RawMessage
	if newCfg != nil {
		b, err := marshalConfigSnapshot(*newCfg)
		if err != nil {
			return err
		}
		newRaw = b
	}

	return u.historyRepo.Create(domain.HistoryEntry{
		EntityType: domain.EntityTypeConfig,
		EntityID:   entityID,
		EntityKey:  entityKey,
		Scope:      scope,
		Action:     action,
		OldValue:   oldRaw,
		NewValue:   newRaw,
		ChangedBy:  adminUserID,
	})
}

func marshalConfigSnapshot(cfg domain.Config) (json.RawMessage, error) {
	payload := map[string]any{
		"id":          cfg.ID,
		"key":         cfg.Key,
		"scope":       cfg.Scope,
		"valueType":   cfg.ValueType,
		"value":       cfg.Value,
		"description": cfg.Description,
	}
	return json.Marshal(payload)
}

func validateConfigValue(valueType string, raw json.RawMessage) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || !json.Valid(trimmed) || string(trimmed) == "null" {
		return domain.ErrConfigValueType
	}

	switch valueType {
	case domain.ValueTypeString:
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return domain.ErrConfigValueType
		}
	case domain.ValueTypeNumber:
		var n float64
		if err := json.Unmarshal(trimmed, &n); err != nil {
			return domain.ErrConfigValueType
		}
	case domain.ValueTypeBoolean:
		var b bool
		if err := json.Unmarshal(trimmed, &b); err != nil {
			return domain.ErrConfigValueType
		}
	case domain.ValueTypeArray:
		var arr []any
		if trimmed[0] != '[' || json.Unmarshal(trimmed, &arr) != nil {
			return domain.ErrConfigValueType
		}
	case domain.ValueTypeJSON:
		var obj map[string]any
		if trimmed[0] != '{' || json.Unmarshal(trimmed, &obj) != nil {
			return domain.ErrConfigValueType
		}
	default:
		return domain.ErrConfigValueType
	}

	return nil
}
