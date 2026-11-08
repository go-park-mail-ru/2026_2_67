package usecase

import (
	"errors"
	"strings"
	"time"
	authDomain "vibe_market/backend/internal/auth/domain"
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

func GetBuyerProfileUsecase(userID int64, buyerRepo domain.BuyerRepo) (GetBuyerProfileOutDTO, error) {
	buyer, user, err := buyerRepo.SelectByUserIDWithUser(userID)
	if errors.Is(err, authDomain.ErrUserIDNotFound) {
		return GetBuyerProfileOutDTO{}, authDomain.ErrUserIDNotFound
	}
	if !errors.Is(err, nil) {
		return GetBuyerProfileOutDTO{}, err
	}
	birthDate := ""
	if buyer.BirthDate != nil {
		birthDate = buyer.BirthDate.String()
	}

	return GetBuyerProfileOutDTO{Name: buyer.Name, Surname: str(buyer.Surname), Telephone: str(buyer.Telephone), Email: user.Email,
		BirthDate: birthDate, Description: str(buyer.Description), AvatarURL: str(buyer.AvatarURL)}, nil
}

type PatchBuyerProfileInDTO struct {
	UserID      int64
	Name        string
	Surname     string
	Telephone   string
	Email       string
	BirthDate   string
	Description string
	AvatarURL   *string
}

func PatchBuyerProfile(dto PatchBuyerProfileInDTO, buyerRepo domain.BuyerRepo) (GetBuyerProfileOutDTO, error) {
	if strings.TrimSpace(dto.Name) == "" || strings.TrimSpace(dto.Email) == "" {
		return GetBuyerProfileOutDTO{}, errors.New("name и email не могут быть пустыми")
	}
	birthDate, err := time.Parse("2006-01-02", dto.BirthDate)
	if err != nil {
		return GetBuyerProfileOutDTO{}, err
	}
	surname, telephone, description := dto.Surname, dto.Telephone, dto.Description
	buyer := domain.Buyer{UserID: dto.UserID, Name: dto.Name, Surname: &surname, BirthDate: &birthDate,
		Description: &description, Telephone: &telephone, AvatarURL: dto.AvatarURL}

	current, _, err := buyerRepo.SelectByUserIDWithUser(dto.UserID)
	if err != nil {
		return GetBuyerProfileOutDTO{}, err
	}
	if dto.AvatarURL == nil {
		buyer.AvatarURL = current.AvatarURL
	}
	ctx, err := buyerRepo.BeginTx()
	if err != nil {
		return GetBuyerProfileOutDTO{}, err
	}
	defer buyerRepo.RollbackTx(ctx)
	if err = buyerRepo.UpdateProfileTx(ctx, buyer, dto.Email); err != nil {
		return GetBuyerProfileOutDTO{}, err
	}
	buyerRepo.CommitTx(ctx)
	return GetBuyerProfileUsecase(dto.UserID, buyerRepo)
}
