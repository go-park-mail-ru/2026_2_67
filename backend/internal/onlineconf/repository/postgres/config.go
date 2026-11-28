package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"vibe_market/backend/internal/onlineconf/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolationCode = "23505"

type ConfigRepoPostgres struct {
	db *sql.DB
}

func NewConfigRepo(db *sql.DB) *ConfigRepoPostgres {
	return &ConfigRepoPostgres{db: db}
}

func (r *ConfigRepoPostgres) List(scope string, query string) ([]domain.Config, error) {
	const sqlQuery = `
		SELECT id, key, scope, value_type, value, description, created_by, updated_by, created_at, updated_at
		FROM oc_config
		WHERE ($1 = '' OR scope = $1)
		  AND ($2 = '' OR key ILIKE '%' || $2 || '%' OR description ILIKE '%' || $2 || '%')
		ORDER BY scope ASC, key ASC;
	`

	rows, err := r.db.Query(sqlQuery, scope, query)
	if err != nil {
		return nil, fmt.Errorf("list oc_config: %w", err)
	}
	defer rows.Close()

	configs := make([]domain.Config, 0)
	for rows.Next() {
		var cfg domain.Config
		var rawValue []byte
		if err := rows.Scan(
			&cfg.ID,
			&cfg.Key,
			&cfg.Scope,
			&cfg.ValueType,
			&rawValue,
			&cfg.Description,
			&cfg.CreatedBy,
			&cfg.UpdatedBy,
			&cfg.CreatedAt,
			&cfg.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan oc_config: %w", err)
		}
		cfg.Value = rawValue
		configs = append(configs, cfg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate oc_config: %w", err)
	}

	return configs, nil
}

func (r *ConfigRepoPostgres) GetByID(id int64) (domain.Config, error) {
	const sqlQuery = `
		SELECT id, key, scope, value_type, value, description, created_by, updated_by, created_at, updated_at
		FROM oc_config
		WHERE id = $1;
	`

	var cfg domain.Config
	var rawValue []byte
	err := r.db.QueryRow(sqlQuery, id).Scan(
		&cfg.ID,
		&cfg.Key,
		&cfg.Scope,
		&cfg.ValueType,
		&rawValue,
		&cfg.Description,
		&cfg.CreatedBy,
		&cfg.UpdatedBy,
		&cfg.CreatedAt,
		&cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Config{}, domain.ErrConfigNotFound
		}
		return domain.Config{}, fmt.Errorf("get oc_config by id: %w", err)
	}
	cfg.Value = rawValue

	return cfg, nil
}

func (r *ConfigRepoPostgres) Create(cfg domain.Config) (domain.Config, error) {
	const sqlQuery = `
		INSERT INTO oc_config (key, scope, value_type, value, description, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING id, key, scope, value_type, value, description, created_by, updated_by, created_at, updated_at;
	`

	var created domain.Config
	var rawValue []byte
	err := r.db.QueryRow(
		sqlQuery,
		cfg.Key,
		cfg.Scope,
		cfg.ValueType,
		[]byte(cfg.Value),
		cfg.Description,
		cfg.CreatedBy,
	).Scan(
		&created.ID,
		&created.Key,
		&created.Scope,
		&created.ValueType,
		&rawValue,
		&created.Description,
		&created.CreatedBy,
		&created.UpdatedBy,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
			return domain.Config{}, domain.ErrConfigAlreadyExists
		}
		return domain.Config{}, fmt.Errorf("create oc_config: %w", err)
	}
	created.Value = rawValue

	return created, nil
}

func (r *ConfigRepoPostgres) Update(cfg domain.Config) (domain.Config, error) {
	const sqlQuery = `
		UPDATE oc_config
		SET value_type = $2,
		    value = $3,
		    description = $4,
		    updated_by = $5,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, key, scope, value_type, value, description, created_by, updated_by, created_at, updated_at;
	`

	var updated domain.Config
	var rawValue []byte
	err := r.db.QueryRow(
		sqlQuery,
		cfg.ID,
		cfg.ValueType,
		[]byte(cfg.Value),
		cfg.Description,
		cfg.UpdatedBy,
	).Scan(
		&updated.ID,
		&updated.Key,
		&updated.Scope,
		&updated.ValueType,
		&rawValue,
		&updated.Description,
		&updated.CreatedBy,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Config{}, domain.ErrConfigNotFound
		}
		return domain.Config{}, fmt.Errorf("update oc_config: %w", err)
	}
	updated.Value = rawValue

	return updated, nil
}

func (r *ConfigRepoPostgres) Delete(id int64) (domain.Config, error) {
	const sqlQuery = `
		DELETE FROM oc_config
		WHERE id = $1
		RETURNING id, key, scope, value_type, value, description, created_by, updated_by, created_at, updated_at;
	`

	var deleted domain.Config
	var rawValue []byte
	err := r.db.QueryRow(sqlQuery, id).Scan(
		&deleted.ID,
		&deleted.Key,
		&deleted.Scope,
		&deleted.ValueType,
		&rawValue,
		&deleted.Description,
		&deleted.CreatedBy,
		&deleted.UpdatedBy,
		&deleted.CreatedAt,
		&deleted.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Config{}, domain.ErrConfigNotFound
		}
		return domain.Config{}, fmt.Errorf("delete oc_config: %w", err)
	}
	deleted.Value = rawValue

	return deleted, nil
}
