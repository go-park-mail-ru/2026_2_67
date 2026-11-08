package usecase

import (
	"errors"
	"strings"
	"time"
	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/usecase"
	"vibe_market/backend/internal/buyers/domain"
)

type BuyerProfileOutDTO struct {
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

func GetBuyerProfileUsecase(userID int64, buyerRepo domain.BuyerRepo) (BuyerProfileOutDTO, error) {
	buyer, user, err := buyerRepo.SelectByUserIDWithUser(userID)
	if errors.Is(err, authDomain.ErrUserIDNotFound) {
		return BuyerProfileOutDTO{}, authDomain.ErrUserIDNotFound
	}
	if !errors.Is(err, nil) {
		return BuyerProfileOutDTO{}, err
	}
	birthDate := ""
	if buyer.BirthDate != nil {
		birthDate = buyer.BirthDate.String()
	}

	return BuyerProfileOutDTO{Name: buyer.Name, Surname: str(buyer.Surname), Telephone: str(buyer.Telephone), Email: user.Email,
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

// WARN: как будто медленно, надо лишние запросы убрать

func PatchBuyerProfile(dto PatchBuyerProfileInDTO, buyerRepo domain.BuyerRepo, userRepo authDomain.UserRepo) (BuyerProfileOutDTO, error) {
	dto.Email = strings.TrimSpace(dto.Email)
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Surname = strings.TrimSpace(dto.Surname)
	dto.Telephone = strings.TrimSpace(dto.Telephone)
	dto.Description = strings.TrimSpace(dto.Description)

	if dto.Email == "" {
		return BuyerProfileOutDTO{}, authDomain.ErrInvalidEmail
	}

	err := usecase.ValidateEmail(dto.Email)
	if err != nil {
		return BuyerProfileOutDTO{}, err
	}
	var birthDate *time.Time
	if dto.BirthDate != "" {
		birthDateVal, err := time.Parse("2006-01-02", dto.BirthDate)
		if err != nil {
			return BuyerProfileOutDTO{}, err
		}
		birthDate = &birthDateVal
	}
	var (
		surname     *string
		telephone   *string
		description *string
	)
	if dto.Surname != "" {
		surname = &dto.Surname
	}
	if dto.Telephone != "" {
		telephone = &dto.Telephone
	}
	if dto.Description != "" {
		description = &dto.Description
	}
	buyer := domain.Buyer{UserID: dto.UserID, Name: dto.Name, Surname: surname, BirthDate: birthDate,
		Description: description, Telephone: telephone, AvatarURL: dto.AvatarURL}
	if buyer.AvatarURL != nil && !strings.HasPrefix(*buyer.AvatarURL, "/") {
		avatarURL := "/" + *buyer.AvatarURL
		buyer.AvatarURL = &avatarURL
	}

	current, _, err := buyerRepo.SelectByUserIDWithUser(dto.UserID)
	if err != nil {
		return BuyerProfileOutDTO{}, err
	}
	if dto.AvatarURL == nil {
		buyer.AvatarURL = current.AvatarURL
	}
	ctx, err := buyerRepo.BeginTx()
	if err != nil {
		return BuyerProfileOutDTO{}, err
	}
	defer buyerRepo.RollbackTx(ctx)

	err = userRepo.UpdateEmailByUserIDTx(ctx, dto.Email, dto.UserID)
	if err != nil {
		return BuyerProfileOutDTO{}, err
	}
	err = buyerRepo.UpdateTx(ctx, buyer)
	if err != nil {
		return BuyerProfileOutDTO{}, err
	}
	buyerRepo.CommitTx(ctx)
	return GetBuyerProfileUsecase(dto.UserID, buyerRepo)
}
