package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"vibe_market/backend/internal/onlineconf/domain"
)

type HistoryRepoPostgres struct {
	db *sql.DB
}

func NewHistoryRepo(db *sql.DB) *HistoryRepoPostgres {
	return &HistoryRepoPostgres{db: db}
}

func nullableJSONB(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return []byte(raw)
}

func (r *HistoryRepoPostgres) Create(entry domain.HistoryEntry) error {
	const sqlQuery = `
		INSERT INTO oc_history (entity_type, entity_id, entity_key, scope, action, old_value, new_value, changed_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`

	_, err := r.db.Exec(
		sqlQuery,
		entry.EntityType,
		entry.EntityID,
		entry.EntityKey,
		entry.Scope,
		entry.Action,
		nullableJSONB(entry.OldValue),
		nullableJSONB(entry.NewValue),
		entry.ChangedBy,
	)
	if err != nil {
		return fmt.Errorf("insert oc_history: %w", err)
	}

	return nil
}

func (r *HistoryRepoPostgres) List(entityType string, entityID int64, limit int) ([]domain.HistoryEntry, error) {
	const sqlQuery = `
		SELECT id, entity_type, entity_id, entity_key, scope, action, old_value, new_value, changed_by, created_at
		FROM oc_history
		WHERE ($1 = '' OR entity_type = $1)
		  AND ($2::BIGINT = 0 OR entity_id = $2)
		ORDER BY created_at DESC, id DESC
		LIMIT $3;
	`

	rows, err := r.db.Query(sqlQuery, entityType, entityID, limit)
	if err != nil {
		return nil, fmt.Errorf("list oc_history: %w", err)
	}
	defer rows.Close()

	entries := make([]domain.HistoryEntry, 0)
	for rows.Next() {
		var entry domain.HistoryEntry
		var oldRaw []byte
		var newRaw []byte
		if err := rows.Scan(
			&entry.ID,
			&entry.EntityType,
			&entry.EntityID,
			&entry.EntityKey,
			&entry.Scope,
			&entry.Action,
			&oldRaw,
			&newRaw,
			&entry.ChangedBy,
			&entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan oc_history: %w", err)
		}
		if len(oldRaw) > 0 {
			entry.OldValue = json.RawMessage(oldRaw)
		}
		if len(newRaw) > 0 {
			entry.NewValue = json.RawMessage(newRaw)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate oc_history: %w", err)
	}

	return entries, nil
}
