package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"vibe_market/backend/internal/onlineconf/domain"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type FeatureFlagRepoPostgres struct {
	db *sql.DB
}

func NewFeatureFlagRepo(db *sql.DB) *FeatureFlagRepoPostgres {
	return &FeatureFlagRepoPostgres{db: db}
}

func normalizeIDSlice(ids []int64) []int64 {
	if ids == nil {
		return []int64{}
	}
	return ids
}

func (r *FeatureFlagRepoPostgres) List(scope string, query string) ([]domain.FeatureFlag, error) {
	const sqlQuery = `
		SELECT id, key, scope, enabled, rollout_percent, whitelist, blacklist, description, created_by, updated_by, created_at, updated_at
		FROM oc_feature_flag
		WHERE ($1 = '' OR scope = $1)
		  AND ($2 = '' OR key ILIKE '%' || $2 || '%' OR description ILIKE '%' || $2 || '%')
		ORDER BY scope ASC, key ASC;
	`

	rows, err := r.db.Query(sqlQuery, scope, query)
	if err != nil {
		return nil, fmt.Errorf("list oc_feature_flag: %w", err)
	}
	defer rows.Close()

	flags := make([]domain.FeatureFlag, 0)
	for rows.Next() {
		var flag domain.FeatureFlag
		var whitelist pgtype.FlatArray[int64]
		var blacklist pgtype.FlatArray[int64]
		if err := rows.Scan(
			&flag.ID,
			&flag.Key,
			&flag.Scope,
			&flag.Enabled,
			&flag.RolloutPercent,
			&whitelist,
			&blacklist,
			&flag.Description,
			&flag.CreatedBy,
			&flag.UpdatedBy,
			&flag.CreatedAt,
			&flag.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan oc_feature_flag: %w", err)
		}
		flag.Whitelist = normalizeIDSlice([]int64(whitelist))
		flag.Blacklist = normalizeIDSlice([]int64(blacklist))
		flags = append(flags, flag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate oc_feature_flag: %w", err)
	}

	return flags, nil
}

func (r *FeatureFlagRepoPostgres) GetByID(id int64) (domain.FeatureFlag, error) {
	const sqlQuery = `
		SELECT id, key, scope, enabled, rollout_percent, whitelist, blacklist, description, created_by, updated_by, created_at, updated_at
		FROM oc_feature_flag
		WHERE id = $1;
	`

	var flag domain.FeatureFlag
	var whitelist pgtype.FlatArray[int64]
	var blacklist pgtype.FlatArray[int64]
	err := r.db.QueryRow(sqlQuery, id).Scan(
		&flag.ID,
		&flag.Key,
		&flag.Scope,
		&flag.Enabled,
		&flag.RolloutPercent,
		&whitelist,
		&blacklist,
		&flag.Description,
		&flag.CreatedBy,
		&flag.UpdatedBy,
		&flag.CreatedAt,
		&flag.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.FeatureFlag{}, domain.ErrFeatureFlagNotFound
		}
		return domain.FeatureFlag{}, fmt.Errorf("get oc_feature_flag by id: %w", err)
	}
	flag.Whitelist = normalizeIDSlice([]int64(whitelist))
	flag.Blacklist = normalizeIDSlice([]int64(blacklist))

	return flag, nil
}

func (r *FeatureFlagRepoPostgres) Create(flag domain.FeatureFlag) (domain.FeatureFlag, error) {
	const sqlQuery = `
		INSERT INTO oc_feature_flag (key, scope, enabled, rollout_percent, whitelist, blacklist, description, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id, key, scope, enabled, rollout_percent, whitelist, blacklist, description, created_by, updated_by, created_at, updated_at;
	`

	var created domain.FeatureFlag
	var whitelist pgtype.FlatArray[int64]
	var blacklist pgtype.FlatArray[int64]
	err := r.db.QueryRow(
		sqlQuery,
		flag.Key,
		flag.Scope,
		flag.Enabled,
		flag.RolloutPercent,
		pgtype.FlatArray[int64](normalizeIDSlice(flag.Whitelist)),
		pgtype.FlatArray[int64](normalizeIDSlice(flag.Blacklist)),
		flag.Description,
		flag.CreatedBy,
	).Scan(
		&created.ID,
		&created.Key,
		&created.Scope,
		&created.Enabled,
		&created.RolloutPercent,
		&whitelist,
		&blacklist,
		&created.Description,
		&created.CreatedBy,
		&created.UpdatedBy,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
			return domain.FeatureFlag{}, domain.ErrFeatureFlagAlreadyExists
		}
		return domain.FeatureFlag{}, fmt.Errorf("create oc_feature_flag: %w", err)
	}
	created.Whitelist = normalizeIDSlice([]int64(whitelist))
	created.Blacklist = normalizeIDSlice([]int64(blacklist))

	return created, nil
}

func (r *FeatureFlagRepoPostgres) Update(flag domain.FeatureFlag) (domain.FeatureFlag, error) {
	const sqlQuery = `
		UPDATE oc_feature_flag
		SET enabled = $2,
		    rollout_percent = $3,
		    whitelist = $4,
		    blacklist = $5,
		    description = $6,
		    updated_by = $7,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, key, scope, enabled, rollout_percent, whitelist, blacklist, description, created_by, updated_by, created_at, updated_at;
	`

	var updated domain.FeatureFlag
	var whitelist pgtype.FlatArray[int64]
	var blacklist pgtype.FlatArray[int64]
	err := r.db.QueryRow(
		sqlQuery,
		flag.ID,
		flag.Enabled,
		flag.RolloutPercent,
		pgtype.FlatArray[int64](normalizeIDSlice(flag.Whitelist)),
		pgtype.FlatArray[int64](normalizeIDSlice(flag.Blacklist)),
		flag.Description,
		flag.UpdatedBy,
	).Scan(
		&updated.ID,
		&updated.Key,
		&updated.Scope,
		&updated.Enabled,
		&updated.RolloutPercent,
		&whitelist,
		&blacklist,
		&updated.Description,
		&updated.CreatedBy,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.FeatureFlag{}, domain.ErrFeatureFlagNotFound
		}
		return domain.FeatureFlag{}, fmt.Errorf("update oc_feature_flag: %w", err)
	}
	updated.Whitelist = normalizeIDSlice([]int64(whitelist))
	updated.Blacklist = normalizeIDSlice([]int64(blacklist))

	return updated, nil
}

func (r *FeatureFlagRepoPostgres) Delete(id int64) (domain.FeatureFlag, error) {
	const sqlQuery = `
		DELETE FROM oc_feature_flag
		WHERE id = $1
		RETURNING id, key, scope, enabled, rollout_percent, whitelist, blacklist, description, created_by, updated_by, created_at, updated_at;
	`

	var deleted domain.FeatureFlag
	var whitelist pgtype.FlatArray[int64]
	var blacklist pgtype.FlatArray[int64]
	err := r.db.QueryRow(sqlQuery, id).Scan(
		&deleted.ID,
		&deleted.Key,
		&deleted.Scope,
		&deleted.Enabled,
		&deleted.RolloutPercent,
		&whitelist,
		&blacklist,
		&deleted.Description,
		&deleted.CreatedBy,
		&deleted.UpdatedBy,
		&deleted.CreatedAt,
		&deleted.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.FeatureFlag{}, domain.ErrFeatureFlagNotFound
		}
		return domain.FeatureFlag{}, fmt.Errorf("delete oc_feature_flag: %w", err)
	}
	deleted.Whitelist = normalizeIDSlice([]int64(whitelist))
	deleted.Blacklist = normalizeIDSlice([]int64(blacklist))

	return deleted, nil
}
