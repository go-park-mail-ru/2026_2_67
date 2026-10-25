package postgres

import (
	"database/sql"
	"errors"
	"vibe_market/backend/internal/auth/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepoPostgres struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepoPostgres {
	return &UserRepoPostgres{db}
}

func (repo *UserRepoPostgres) Insert(email string, passwordHash string) (domain.User, error) {
	row := repo.db.QueryRow(
		"INSERT INTO \"user\" (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash, created_at, updated_at;",
		email, passwordHash,
	)
	user := domain.User{}

	if row.Err() != nil {
		var pgErr *pgconn.PgError
		if errors.As(row.Err(), &pgErr) && pgErr.Code == "23505" {
			return user, domain.ErrEmailAlreadyExists
		}
		return user, row.Err()
	}

	err := row.Scan(&user.UserID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (repo *UserRepoPostgres) DeleteByUserID(userID int64) error {
	row := repo.db.QueryRow("DELETE FROM \"user\" WHERE id = $1 RETURNING id;", userID)
	if row.Err() != nil {
		return row.Err()
	}

	var deletedID int64
	err := row.Scan(&deletedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrUserIDNotFound
		}
		return err
	}
	return nil
}

func (repo *UserRepoPostgres) SelectByEmail(email string) (domain.User, error) {
	row := repo.db.QueryRow(
		"SELECT id, email, password_hash, created_at, updated_at FROM \"user\" WHERE email = $1;",
		email,
	)
	user := domain.User{}

	if row.Err() != nil {
		return user, row.Err()
	}

	err := row.Scan(&user.UserID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, domain.ErrEmailNotFound
		}
		return user, err
	}

	return user, nil
}

func (repo *UserRepoPostgres) SelectByUserID(userID int64) (domain.User, error) {
	row := repo.db.QueryRow("SELECT id, email, password_hash, created_at, updated_at FROM \"user\" WHERE id = $1", userID)
	user := domain.User{}

	if row.Err() != nil {
		return user, row.Err()
	}

	err := row.Scan(&user.UserID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, domain.ErrUserIDNotFound
		}
		return user, err
	}
	return user, nil
}
