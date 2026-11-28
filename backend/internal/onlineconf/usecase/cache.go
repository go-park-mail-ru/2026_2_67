package usecase

import (
	"context"
	"sync"
	"time"
	"vibe_market/backend/internal/onlineconf/domain"

	"go.uber.org/zap"
)

const defaultCacheRefreshInterval = 10 * time.Second

type RuntimeCache struct {
	configRepo domain.ConfigRepo
	flagRepo   domain.FeatureFlagRepo
	logger     *zap.Logger

	mu      sync.RWMutex
	configs []domain.Config
	flags   []domain.FeatureFlag
}

func NewRuntimeCache(configRepo domain.ConfigRepo, flagRepo domain.FeatureFlagRepo, logger *zap.Logger) *RuntimeCache {
	return &RuntimeCache{
		configRepo: configRepo,
		flagRepo:   flagRepo,
		logger:     logger,
		configs:    make([]domain.Config, 0),
		flags:      make([]domain.FeatureFlag, 0),
	}
}

func (c *RuntimeCache) Start(ctx context.Context) {
	if err := c.Refresh(); err != nil && c.logger != nil {
		c.logger.Warn("не удалось выполнить первичную загрузку кеша onlineconf", zap.Error(err))
	}

	ticker := time.NewTicker(defaultCacheRefreshInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.Refresh(); err != nil && c.logger != nil {
					c.logger.Warn("не удалось обновить кеш onlineconf", zap.Error(err))
				}
			}
		}
	}()
}

func (c *RuntimeCache) Refresh() error {
	configs, err := c.configRepo.List("", "")
	if err != nil {
		return err
	}

	flags, err := c.flagRepo.List("", "")
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.configs = configs
	c.flags = flags
	c.mu.Unlock()

	return nil
}

func (c *RuntimeCache) Invalidate() {
	if err := c.Refresh(); err != nil && c.logger != nil {
		c.logger.Warn("не удалось обновить кеш onlineconf после записи", zap.Error(err))
	}
}

func (c *RuntimeCache) Snapshot() ([]domain.Config, []domain.FeatureFlag) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	configsCopy := make([]domain.Config, len(c.configs))
	copy(configsCopy, c.configs)

	flagsCopy := make([]domain.FeatureFlag, len(c.flags))
	copy(flagsCopy, c.flags)

	return configsCopy, flagsCopy
}
