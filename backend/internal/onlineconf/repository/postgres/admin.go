package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"vibe_market/backend/internal/onlineconf/domain"
)

type AdminRepoPostgres struct {
	db *sql.DB
}

func NewAdminRepo(db *sql.DB) *AdminRepoPostgres {
	return &AdminRepoPostgres{db: db}
}

func (r *AdminRepoPostgres) GetByUserID(userID int64) (domain.Admin, error) {
	const query = `
		SELECT user_id, role, created_at
		FROM oc_admin
		WHERE user_id = $1;
	`

	var admin domain.Admin
	err := r.db.QueryRow(query, userID).Scan(&admin.UserID, &admin.Role, &admin.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Admin{}, domain.ErrAdminForbidden
		}
		return domain.Admin{}, fmt.Errorf("select oc_admin by user_id: %w", err)
	}

	return admin, nil
}
