package domain

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	EntityTypeConfig = "config"
	EntityTypeFlag   = "flag"

	HistoryActionCreate = "create"
	HistoryActionUpdate = "update"
	HistoryActionDelete = "delete"
)

var (
	ErrHistoryEntityTypeInvalid = errors.New("некорректный тип сущности истории")
	ErrHistoryLimitInvalid      = errors.New("некорректный лимит истории")
)

type HistoryEntry struct {
	ID         int64
	EntityType string
	EntityID   int64
	EntityKey  string
	Scope      string
	Action     string
	OldValue   json.RawMessage
	NewValue   json.RawMessage
	ChangedBy  int64
	CreatedAt  time.Time
}

type ListHistoryInDTO struct {
	EntityType string
	EntityID   int64
	Limit      int
}

type HistoryRepo interface {
	Create(entry HistoryEntry) error
	List(entityType string, entityID int64, limit int) ([]HistoryEntry, error)
}

type HistoryUsecase interface {
	List(dto ListHistoryInDTO) ([]HistoryEntry, error)
}
