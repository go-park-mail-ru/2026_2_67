package usecase

import (
	"encoding/json"
	"strings"
	"vibe_market/backend/internal/onlineconf/domain"
)

type FeatureFlagUsecase struct {
	flagRepo    domain.FeatureFlagRepo
	historyRepo domain.HistoryRepo
	cache       *RuntimeCache
}

func NewFeatureFlagUsecase(flagRepo domain.FeatureFlagRepo, historyRepo domain.HistoryRepo, cache *RuntimeCache) *FeatureFlagUsecase {
	return &FeatureFlagUsecase{
		flagRepo:    flagRepo,
		historyRepo: historyRepo,
		cache:       cache,
	}
}

func (u *FeatureFlagUsecase) List(scope string, query string) ([]domain.FeatureFlag, error) {
	return u.flagRepo.List(strings.TrimSpace(scope), strings.TrimSpace(query))
}

func (u *FeatureFlagUsecase) Create(dto domain.CreateFeatureFlagInDTO) (domain.FeatureFlag, error) {
	key := strings.TrimSpace(dto.Key)
	scope := strings.TrimSpace(dto.Scope)
	if !domain.ValidateIdentifier(key) {
		return domain.FeatureFlag{}, domain.ErrKeyInvalid
	}
	if !domain.ValidateIdentifier(scope) {
		return domain.FeatureFlag{}, domain.ErrScopeInvalid
	}
	if dto.RolloutPercent < 0 || dto.RolloutPercent > 100 {
		return domain.FeatureFlag{}, domain.ErrFeatureFlagRolloutInvalid
	}

	created, err := u.flagRepo.Create(domain.FeatureFlag{
		Key:            key,
		Scope:          scope,
		Enabled:        dto.Enabled,
		RolloutPercent: dto.RolloutPercent,
		Whitelist:      normalizeSlice(dto.Whitelist),
		Blacklist:      normalizeSlice(dto.Blacklist),
		Description:    dto.Description,
		CreatedBy:      dto.AdminUserID,
		UpdatedBy:      dto.AdminUserID,
	})
	if err != nil {
		return domain.FeatureFlag{}, err
	}

	if err := u.recordHistory(domain.HistoryActionCreate, created.ID, created.Key, created.Scope, nil, &created, dto.AdminUserID); err != nil {
		return domain.FeatureFlag{}, err
	}

	if u.cache != nil {
		u.cache.Invalidate()
	}

	return created, nil
}

func (u *FeatureFlagUsecase) Update(dto domain.UpdateFeatureFlagInDTO) (domain.FeatureFlag, error) {
	if dto.ID <= 0 {
		return domain.FeatureFlag{}, domain.ErrFeatureFlagInvalid
	}
	if dto.RolloutPercent < 0 || dto.RolloutPercent > 100 {
		return domain.FeatureFlag{}, domain.ErrFeatureFlagRolloutInvalid
	}

	oldFlag, err := u.flagRepo.GetByID(dto.ID)
	if err != nil {
		return domain.FeatureFlag{}, err
	}

	updated, err := u.flagRepo.Update(domain.FeatureFlag{
		ID:             dto.ID,
		Enabled:        dto.Enabled,
		RolloutPercent: dto.RolloutPercent,
		Whitelist:      normalizeSlice(dto.Whitelist),
		Blacklist:      normalizeSlice(dto.Blacklist),
		Description:    dto.Description,
		UpdatedBy:      dto.AdminUserID,
	})
	if err != nil {
		return domain.FeatureFlag{}, err
	}

	if err := u.recordHistory(domain.HistoryActionUpdate, updated.ID, updated.Key, updated.Scope, &oldFlag, &updated, dto.AdminUserID); err != nil {
		return domain.FeatureFlag{}, err
	}

	if u.cache != nil {
		u.cache.Invalidate()
	}

	return updated, nil
}

func (u *FeatureFlagUsecase) Delete(id int64, adminUserID int64) error {
	if id <= 0 {
		return domain.ErrFeatureFlagNotFound
	}

	deleted, err := u.flagRepo.Delete(id)
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

func (u *FeatureFlagUsecase) Evaluate(id int64, userID int64) (domain.EvaluateFeatureFlagOutDTO, error) {
	flag, err := u.flagRepo.GetByID(id)
	if err != nil {
		return domain.EvaluateFeatureFlagOutDTO{}, err
	}

	return domain.EvaluateFeatureFlag(flag, userID), nil
}

func (u *FeatureFlagUsecase) recordHistory(
	action string,
	entityID int64,
	entityKey string,
	scope string,
	oldFlag *domain.FeatureFlag,
	newFlag *domain.FeatureFlag,
	adminUserID int64,
) error {
	var oldRaw json.RawMessage
	if oldFlag != nil {
		b, err := marshalFlagSnapshot(*oldFlag)
		if err != nil {
			return err
		}
		oldRaw = b
	}

	var newRaw json.RawMessage
	if newFlag != nil {
		b, err := marshalFlagSnapshot(*newFlag)
		if err != nil {
			return err
		}
		newRaw = b
	}

	return u.historyRepo.Create(domain.HistoryEntry{
		EntityType: domain.EntityTypeFlag,
		EntityID:   entityID,
		EntityKey:  entityKey,
		Scope:      scope,
		Action:     action,
		OldValue:   oldRaw,
		NewValue:   newRaw,
		ChangedBy:  adminUserID,
	})
}

func marshalFlagSnapshot(flag domain.FeatureFlag) (json.RawMessage, error) {
	payload := map[string]any{
		"id":             flag.ID,
		"key":            flag.Key,
		"scope":          flag.Scope,
		"enabled":        flag.Enabled,
		"rolloutPercent": flag.RolloutPercent,
		"whitelist":      normalizeSlice(flag.Whitelist),
		"blacklist":      normalizeSlice(flag.Blacklist),
		"description":    flag.Description,
	}
	return json.Marshal(payload)
}

func normalizeSlice(ids []int64) []int64 {
	if ids == nil {
		return []int64{}
	}
	return ids
}
