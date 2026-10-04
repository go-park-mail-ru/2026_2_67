package auth

import "time"

type RoleType string

const (
	RoleBuyer        RoleType = "buyer"
	RoleSeller       RoleType = "seller"
	RollePickupPoint RoleType = "pickup_point"
)

type User struct {
	UserID       int64    `json:"userId"`
	Login        string   `json:"login"`
	Role         RoleType `json:"role"`
	passwordHash string
}

type RefreshToken struct {
	UserID    int64
	TokenHash string
	IsRevoked bool
	ExpiresAt time.Time
}
