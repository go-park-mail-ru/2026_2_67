package postgres

import (
	"database/sql"
	"vibe_market/backend/internal/auth/domain"
)

type UserRepoPostgres struct {
	db *sql.DB
}

const postgresMaxConnections = 20

func NewUserRepo(connURL string) (*UserRepoPostgres, error) {
	userRepo := &UserRepoPostgres{}

	db, err := sql.Open("pgx", connURL)
	if err != nil {
		return userRepo, err
	}
	err = db.Ping()
	if err != nil {
		return userRepo, err
	}

	userRepo.db = db
	userRepo.db.SetMaxOpenConns(postgresMaxConnections)

	return userRepo, nil
}

func (repo *UserRepoPostgres) Insert(email string, passwordHash string) (domain.User, error) {
	row := repo.db.QueryRow(
		"INSERT INTO \"user\" (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash, created_at, updated_at;",
		email, passwordHash,
	)
	user := domain.User{}

	if row.Err() != nil {
		return user, row.Err()
	}

	err := row.Scan(&user.UserID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (repo *UserRepoPostgres) DeleteByUserID(userID int64) error {
	row := repo.db.QueryRow("DELETE FROM \"user\" WHERE id = $1 ", userID)
	return row.Err()
}

func (repo *UserRepoPostgres) SelectByEmail(email string) (domain.User, error) {
	row := repo.db.QueryRow(
		"SELECT id, email, password_hash, created_at, updated_at FROM \"user\" WHERE email = $1",
		email,
	)
	user := domain.User{}

	if row.Err() != nil {
		return user, row.Err()
	}

	err := row.Scan(&user.UserID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (repo *UserRepoPostgres) SelectByUserID(userID int64) (domain.User, error) {
	row := repo.db.QueryRow("SELECT id, email, password_hash, created_at, updated_at FROM \"user\" WHERE id = $1", userID)
	user := domain.User{}

	if row.Err() != nil {
		return user, row.Err()
	}

	err := row.Scan(&user.UserID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}
