package domain

import (
	"errors"
	"time"
)

const (
	AdminRoleSuperadmin = "superadmin"
	AdminRoleAdmin      = "admin"
)

var (
	ErrAdminForbidden = errors.New("нет доступа к панели администратора")
)

type Admin struct {
	UserID    int64
	Role      string
	CreatedAt time.Time
}

type AdminRepo interface {
	GetByUserID(userID int64) (Admin, error)
}

type AdminUsecase interface {
	GetByUserID(userID int64) (Admin, error)
}
