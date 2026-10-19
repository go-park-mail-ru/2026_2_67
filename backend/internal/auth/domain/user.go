package domain

import (
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
	Insert(email string, password string) (User, error)
	Truncate(userID int64) error
	SelectByEmail(email string) (User, bool)
	SelectByUserID(userID int64) (User, bool)
	TruncateByUserID(userID int64) bool
}

var (
	ErrEmailAlreadyExists = errors.New("email занят другим пользователем")
)
