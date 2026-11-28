package usecase

import (
	"vibe_market/backend/internal/onlineconf/domain"
)

type AdminUsecase struct {
	adminRepo domain.AdminRepo
}

func NewAdminUsecase(adminRepo domain.AdminRepo) *AdminUsecase {
	return &AdminUsecase{adminRepo: adminRepo}
}

func (u *AdminUsecase) GetByUserID(userID int64) (domain.Admin, error) {
	if userID <= 0 {
		return domain.Admin{}, domain.ErrAdminForbidden
	}
	return u.adminRepo.GetByUserID(userID)
}
