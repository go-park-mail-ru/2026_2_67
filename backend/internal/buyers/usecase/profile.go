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

func str(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

var (
	ErrBuyerNotFound = errors.New("покупатель не найден")
)

func GetBuyerProfileUsecase(userID int64, buyerRepo domain.BuyerRepo) (GetBuyerProfileOutDTO, error) {
	buyer, user, err := buyerRepo.SelectByUserIDWithUser(userID)
	if err != nil {
		return GetBuyerProfileOutDTO{}, ErrBuyerNotFound
	}
	birthDate := ""
	if buyer.BirthDate != nil {
		birthDate = buyer.BirthDate.String()
	}

	return GetBuyerProfileOutDTO{Name: buyer.Name, Surname: str(buyer.Surname), Telephone: str(buyer.Telephone), Email: user.Email,
		BirthDate: birthDate, Description: str(buyer.Description), AvatarURL: str(buyer.AvatarURL)}, nil
}
