package usecase

import (
	"strings"
	"vibe_market/backend/internal/onlineconf/domain"
)

const (
	defaultHistoryLimit = 50
	maxHistoryLimit     = 500
)

type HistoryUsecase struct {
	historyRepo domain.HistoryRepo
}

func NewHistoryUsecase(historyRepo domain.HistoryRepo) *HistoryUsecase {
	return &HistoryUsecase{historyRepo: historyRepo}
}

func (u *HistoryUsecase) List(dto domain.ListHistoryInDTO) ([]domain.HistoryEntry, error) {
	entityType := strings.TrimSpace(dto.EntityType)
	if entityType != "" && entityType != domain.EntityTypeConfig && entityType != domain.EntityTypeFlag {
		return nil, domain.ErrHistoryEntityTypeInvalid
	}

	if dto.EntityID < 0 {
		return nil, domain.ErrHistoryLimitInvalid
	}

	limit := dto.Limit
	if limit == 0 {
		limit = defaultHistoryLimit
	} else if limit < 0 || limit > maxHistoryLimit {
		return nil, domain.ErrHistoryLimitInvalid
	}

	return u.historyRepo.List(entityType, dto.EntityID, limit)
}
