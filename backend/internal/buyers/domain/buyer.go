package domain

import (
	"time"
	"vibe_market/backend/internal/auth/domain"
)

type Buyer struct {
	domain.User
	Name        string
	Surname     string
	BirthDate   time.Time
	Description string
	AvatarURL   string
	Telephone   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BuyerRepo interface {
	SelectByUserID(userID int64) (Buyer, bool)
	Insert(buyer Buyer) bool
}
