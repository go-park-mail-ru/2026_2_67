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
	Insert(email string, passwordHash string) (User, error)
	DeleteByUserID(userID int64) error
	SelectByEmail(email string) (User, bool)
	SelectByUserID(userID int64) (User, bool)
}

var (
	ErrEmailAlreadyExists = errors.New("email занят другим пользователем")
)
