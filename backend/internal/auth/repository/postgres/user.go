package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"vibe_market/backend/internal/auth/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepoPostgres struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepoPostgres {
	return &UserRepoPostgres{db}
}

func (repo *UserRepoPostgres) BeginTx() (context.Context, error) {
	tx, err := repo.db.BeginTx(context.Background(), nil)
	if err != nil {
		return context.Background(), err
	}
	ctx := context.WithValue(context.Background(), domain.ContextTxKey, tx)
	return ctx, nil
}

func (repo *UserRepoPostgres) InsertTx(ctx context.Context, email string, passwordHash string) (domain.User, error) {
	tx, ok := ctx.Value(domain.ContextTxKey).(*sql.Tx)
	if !ok {
		return domain.User{}, fmt.Errorf("'%s': %s", domain.ContextTxKey, domain.ErrContextKey.Error())
	}
	row := tx.QueryRow(
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

func (repo *UserRepoPostgres) RollbackTx(ctx context.Context) {
	tx, ok := ctx.Value(domain.ContextTxKey).(*sql.Tx)
	if !ok {
		return
	}
	tx.Rollback()
}

func (repo *UserRepoPostgres) CommitTx(ctx context.Context) {
	tx, ok := ctx.Value(domain.ContextTxKey).(*sql.Tx)
	if !ok {
		return
	}
	tx.Commit()
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
	row := repo.db.QueryRow("SELECT id, email, password_hash, created_at, updated_at FROM \"user\" WHERE id = $1;", userID)
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
