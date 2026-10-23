package domain

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AccessTokenTTL time.Duration = 15 * time.Minute

type AccessTokenPayload struct {
	User
	jwt.RegisteredClaims
}

const RefreshTokenTTL time.Duration = 7 * 24 * time.Hour

type RefreshToken struct {
	UserID     int64
	TokenHash  string
	DeviceName string
	IsRevoked  bool
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

type JwtAuthMeta struct {
	JwtIssuer string
	JwtSecret []byte
}

var (
	ErrRefreshTokenNotFound = errors.New("refresh_token не найден")
	ErrRefreshTokenRevoked  = errors.New("refresh_token отозван")
	ErrRefreshTokenExpired  = errors.New("refresh_token истёк")
)

type RefreshTokenRepo interface {
	Insert(t RefreshToken) error
	Revoke(userID int64, deviceName string) error
	SelectByHashAndDeviceName(hash string, deviceName string) (RefreshToken, error)
}
