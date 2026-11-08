package domain

import (
	"context"
	"errors"
	"time"
)

type User struct {
	UserID       int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepo interface {
	BeginTx() (context.Context, error)
	InsertTx(ctx context.Context, email string, passwordHash string) (User, error)
	RollbackTx(ctx context.Context)
	CommitTx(ctx context.Context)
	UpdateEmailByUserIDTx(ctx context.Context, email string, userID int64) error
	UpdatePasswordByUserID(userID int64, passwordHash string) error
	SelectByEmail(email string) (User, error)
	SelectByUserID(userID int64) (User, error)
}

type contextKey string

const ContextTxKey contextKey = "tx"

var (
	ErrEmailAlreadyExists = errors.New("email занят другим пользователем")
	ErrEmailNotFound      = errors.New("user с таким пользователем не найден")
	ErrUserIDNotFound     = errors.New("ID пользователя не найдено")
	ErrContextKey         = errors.New("ключ в context не задан")
)
