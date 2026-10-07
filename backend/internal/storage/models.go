package storage

import "time"

type RoleType string

const (
	RoleBuyer        RoleType = "buyer"
	RoleSeller       RoleType = "seller"
	RollePickupPoint RoleType = "pickup_point"
)

type User struct {
	UserID       int64  `json:"userId"`
	Login        string `json:"login"`
	Email        string `json:"email"`
	PasswordHash string
}

type RefreshToken struct {
	UserID    int64
	TokenHash string
	IsRevoked bool
	ExpiresAt time.Time
}
