package usecase

import (
	"encoding/json"
	"vibe_market/backend/internal/onlineconf/domain"
)

const (
	scopeFrontend = "frontend"
	scopeCommon   = "common"
)

type RuntimeUsecase struct {
	cache *RuntimeCache
}

func NewRuntimeUsecase(cache *RuntimeCache) *RuntimeUsecase {
	return &RuntimeUsecase{cache: cache}
}

func (u *RuntimeUsecase) FrontendRuntime(userID int64) domain.RuntimeFrontendOutDTO {
	configs, flags := u.cache.Snapshot()

	res := domain.RuntimeFrontendOutDTO{
		Configs: make(map[string]json.RawMessage),
		Flags:   make(map[string]bool),
	}

	// Сначала common, затем frontend (чтобы frontend при совпадении ключа имел приоритет).
	for _, cfg := range configs {
		if cfg.Scope == scopeCommon {
			res.Configs[cfg.Key] = cfg.Value
		}
	}
	for _, cfg := range configs {
		if cfg.Scope == scopeFrontend {
			res.Configs[cfg.Key] = cfg.Value
		}
	}

	for _, flag := range flags {
		if flag.Scope == scopeCommon {
			res.Flags[flag.Key] = domain.EvaluateFeatureFlag(flag, userID).Enabled
		}
	}
	for _, flag := range flags {
		if flag.Scope == scopeFrontend {
			res.Flags[flag.Key] = domain.EvaluateFeatureFlag(flag, userID).Enabled
		}
	}

	return res
}

func (u *RuntimeUsecase) Snapshot() domain.SnapshotOutDTO {
	configs, flags := u.cache.Snapshot()

	res := domain.SnapshotOutDTO{
		Configs: make([]domain.SnapshotConfigOutDTO, 0, len(configs)),
		Flags:   make([]domain.SnapshotFlagOutDTO, 0, len(flags)),
	}

	for _, cfg := range configs {
		res.Configs = append(res.Configs, domain.SnapshotConfigOutDTO{
			Scope:     cfg.Scope,
			Key:       cfg.Key,
			ValueType: cfg.ValueType,
			Value:     cfg.Value,
		})
	}

	for _, flag := range flags {
		res.Flags = append(res.Flags, domain.SnapshotFlagOutDTO{
			Scope:          flag.Scope,
			Key:            flag.Key,
			Enabled:        flag.Enabled,
			RolloutPercent: flag.RolloutPercent,
			Whitelist:      normalizeSlice(flag.Whitelist),
			Blacklist:      normalizeSlice(flag.Blacklist),
		})
	}

	return res
}

func (u *RuntimeUsecase) Invalidate() {
	u.cache.Invalidate()
}
