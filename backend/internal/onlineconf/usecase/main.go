package usecase

import (
	"vibe_market/backend/internal/onlineconf/domain"

	"go.uber.org/zap"
)

type Usecases struct {
	Admin       domain.AdminUsecase
	Config      domain.ConfigUsecase
	FeatureFlag domain.FeatureFlagUsecase
	History     domain.HistoryUsecase
	Runtime     domain.RuntimeUsecase
	Cache       *RuntimeCache
}

func NewUsecases(
	adminRepo domain.AdminRepo,
	configRepo domain.ConfigRepo,
	flagRepo domain.FeatureFlagRepo,
	historyRepo domain.HistoryRepo,
	logger *zap.Logger,
) *Usecases {
	cache := NewRuntimeCache(configRepo, flagRepo, logger)
	return &Usecases{
		Admin:       NewAdminUsecase(adminRepo),
		Config:      NewConfigUsecase(configRepo, historyRepo, cache),
		FeatureFlag: NewFeatureFlagUsecase(flagRepo, historyRepo, cache),
		History:     NewHistoryUsecase(historyRepo),
		Runtime:     NewRuntimeUsecase(cache),
		Cache:       cache,
	}
}
