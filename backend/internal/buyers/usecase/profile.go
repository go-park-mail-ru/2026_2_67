package usecase

import (
	"context"
	"errors"
	"image"
	"io"
	"path/filepath"
	"strings"
	"time"
	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/usecase"
	"vibe_market/backend/internal/buyers/domain"
)

func str(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

type BuyerOutDTO struct {
	Name        string
	Surname     string
	Telephone   string
	Email       string
	BirthDate   string
	Description string
	AvatarURL   string
}

func (u *BuyerUsecaseImpl) Buyer(userID int64) (BuyerOutDTO, error) {
	buyer, user, err := u.buyerRepo.SelectByUserIDWithUser(userID)
	if errors.Is(err, authDomain.ErrUserIDNotFound) {
		return BuyerOutDTO{}, authDomain.ErrUserIDNotFound
	}
	if err != nil {
		return BuyerOutDTO{}, err
	}
	birthDate := ""
	if buyer.BirthDate != nil {
		birthDate = buyer.BirthDate.String()
	}

	return BuyerOutDTO{Name: buyer.Name, Surname: str(buyer.Surname), Telephone: str(buyer.Telephone), Email: user.Email,
		BirthDate: birthDate, Description: str(buyer.Description), AvatarURL: str(buyer.AvatarURL)}, nil
}

type PatchBuyerAvatarInDTO struct {
	Ctx         context.Context
	SrcFileName string
	File        io.ReadSeeker
	AvatarURL   string
	Size        int64
}

type PatchBuyerInDTO struct {
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

func (u *BuyerUsecaseImpl) PatchBuyer(dto PatchBuyerInDTO) (BuyerOutDTO, error) {
	dto.Email = strings.TrimSpace(dto.Email)
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Surname = strings.TrimSpace(dto.Surname)
	dto.Telephone = strings.TrimSpace(dto.Telephone)
	dto.Description = strings.TrimSpace(dto.Description)

	if dto.Email == "" {
		return BuyerOutDTO{}, authDomain.ErrInvalidEmail
	}

	err := usecase.ValidateEmail(dto.Email)
	if err != nil {
		return BuyerOutDTO{}, err
	}
	var birthDate *time.Time
	if dto.BirthDate != "" {
		birthDateVal, err := time.Parse("2006-01-02", dto.BirthDate)
		if err != nil {
			return BuyerOutDTO{}, err
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

	current, _, err := u.buyerRepo.SelectByUserIDWithUser(dto.UserID)
	if err != nil {
		return BuyerOutDTO{}, err
	}
	if dto.AvatarURL == nil {
		buyer.AvatarURL = current.AvatarURL
	}
	ctx, err := u.buyerRepo.BeginTx()
	if err != nil {
		return BuyerOutDTO{}, err
	}
	defer u.buyerRepo.RollbackTx(ctx)

	err = u.userRepo.UpdateEmailByUserIDTx(ctx, dto.Email, dto.UserID)
	if err != nil {
		return BuyerOutDTO{}, err
	}
	err = u.buyerRepo.UpdateTx(ctx, buyer)
	if err != nil {
		return BuyerOutDTO{}, err
	}
	u.buyerRepo.CommitTx(ctx)
	return u.Buyer(dto.UserID)
}

func (u *BuyerUsecaseImpl) PatchBuyerAvatar(dto PatchBuyerAvatarInDTO) error {
	if dto.File == nil {
		return nil
	}
	dto.SrcFileName = filepath.Base(dto.SrcFileName)
	if dto.SrcFileName == "." || dto.SrcFileName == "" || strings.Contains(dto.SrcFileName, "\\") {
		return domain.ErrAvatarNameInvalid
	}
	extension := strings.ToLower(filepath.Ext(dto.SrcFileName))
	formats := map[string]string{".jpg": "jpeg", ".jpeg": "jpeg", ".png": "png", ".gif": "gif"}
	expectedFormat, ok := formats[extension]
	if !ok {
		return domain.ErrAvatarNameInvalid
	}
	config, format, err := image.DecodeConfig(dto.File)
	if err != nil || format != expectedFormat || config.Width <= 0 || config.Height <= 0 || int64(config.Width)*int64(config.Height) > 40_000_000 {
		return domain.ErrAvatarFormatInvalid
	}
	if _, err := dto.File.Seek(0, io.SeekStart); err != nil {
		return domain.ErrAvatarFormatInvalid
	}
	if _, format, err = image.Decode(dto.File); err != nil || format != expectedFormat {
		return domain.ErrAvatarFormatInvalid
	}
	if _, err := dto.File.Seek(0, io.SeekStart); err != nil {
		return domain.ErrAvatarFormatInvalid
	}
	contentType := "image/" + format
	key := strings.TrimLeft(dto.AvatarURL, "/"+u.avatarRepo.Bucket()+"/")
	if err := u.avatarRepo.Put(dto.Ctx, key, dto.File, dto.Size, contentType); err != nil {
		return err
	}
	return nil
}
