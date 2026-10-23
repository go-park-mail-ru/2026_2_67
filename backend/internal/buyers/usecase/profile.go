package usecase

import (
	"errors"
	"vibe_market/backend/internal/buyers/domain"
)

type GetBuyerProfileOutDTO struct {
	Name        string
	Surname     string
	Telephone   string
	Email       string
	BirthDate   string
	Description string
	AvatarURL   string
}

var (
	ErrBuyerNotFound = errors.New("покупатель не найден")
)

func GetBuyerProfileUsecase(userID int64, buyerRepo domain.BuyerRepo) (GetBuyerProfileOutDTO, error) {
	buyer, user, err := buyerRepo.SelectByUserIDWithUser(userID)
	if err != nil {
		return GetBuyerProfileOutDTO{}, ErrBuyerNotFound
	}

	return GetBuyerProfileOutDTO{Name: buyer.Name, Surname: *buyer.Surname, Telephone: *buyer.Telephone, Email: user.Email,
		BirthDate: buyer.BirthDate.String(), Description: *buyer.Description, AvatarURL: *buyer.AvatarURL}, nil
}
