package domain

import (
	"context"
	"errors"
	"io"
	"time"
	"vibe_market/backend/internal/auth/domain"
)

type Buyer struct {
	UserID      int64
	Name        string
	Surname     *string
	BirthDate   *time.Time
	Description *string
	AvatarURL   *string
	Telephone   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	ErrUserIDAlreadyExists = errors.New("профиль с таким userID уже существует")
)

type BuyerRepo interface {
	SelectByUserID(userID int64) (Buyer, error)
	SelectByUserIDWithUser(userID int64) (Buyer, domain.User, error)
	BeginTx() (context.Context, error)
	RollbackTx(ctx context.Context)
	InsertTx(ctx context.Context, buyer Buyer) error
	UpdateProfileTx(ctx context.Context, buyer Buyer, email string) error
	CommitTx(ctx context.Context)
}

type BuyerAvatarRepo interface {
	Put(ctx context.Context, objectName string, body io.Reader, size int64, contentType string) error
}
