package usecase

import (
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

func (u *BuyerUsecase) Buyer(userID int64) (domain.BuyerOutDTO, error) {
	buyer, user, err := u.buyerRepo.SelectByUserIDWithUser(userID)
	if errors.Is(err, authDomain.ErrUserIDNotFound) {
		return domain.BuyerOutDTO{}, authDomain.ErrUserIDNotFound
	}
	if err != nil {
		return domain.BuyerOutDTO{}, err
	}
	birthDate := ""
	if buyer.BirthDate != nil {
		birthDate = buyer.BirthDate.String()
	}

	return domain.BuyerOutDTO{Name: buyer.Name, Surname: str(buyer.Surname), Telephone: str(buyer.Telephone), Email: user.Email,
		BirthDate: birthDate, Description: str(buyer.Description), AvatarURL: str(buyer.AvatarURL)}, nil
}

// WARN: как будто медленно, надо лишние запросы убрать

func (u *BuyerUsecase) PatchBuyer(dto domain.PatchBuyerInDTO) (domain.BuyerOutDTO, error) {
	dto.Email = strings.TrimSpace(dto.Email)
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Surname = strings.TrimSpace(dto.Surname)
	dto.Telephone = strings.TrimSpace(dto.Telephone)
	dto.Description = strings.TrimSpace(dto.Description)

	if dto.Email == "" {
		return domain.BuyerOutDTO{}, authDomain.ErrInvalidEmail
	}

	err := usecase.ValidateEmail(dto.Email)
	if err != nil {
		return domain.BuyerOutDTO{}, err
	}
	var birthDate *time.Time
	if dto.BirthDate != "" {
		birthDateVal, err := time.Parse("2006-01-02", dto.BirthDate)
		if err != nil {
			return domain.BuyerOutDTO{}, err
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
		return domain.BuyerOutDTO{}, err
	}
	if dto.AvatarURL == nil {
		buyer.AvatarURL = current.AvatarURL
	}
	ctx, err := u.buyerRepo.BeginTx()
	if err != nil {
		return domain.BuyerOutDTO{}, err
	}
	defer u.buyerRepo.RollbackTx(ctx)

	err = u.userRepo.UpdateEmailByUserIDTx(ctx, dto.Email, dto.UserID)
	if err != nil {
		return domain.BuyerOutDTO{}, err
	}
	err = u.buyerRepo.UpdateTx(ctx, buyer)
	if err != nil {
		return domain.BuyerOutDTO{}, err
	}
	u.buyerRepo.CommitTx(ctx)
	return u.Buyer(dto.UserID)
}

func (u *BuyerUsecase) PatchBuyerAvatar(dto domain.PatchBuyerAvatarInDTO) error {
	if dto.File == nil {
		return nil
	}
	dto.SrcFileName = filepath.Base(dto.SrcFileName)
	if dto.SrcFileName == "." || dto.SrcFileName == "" || strings.Contains(dto.SrcFileName, "\\") {
		return domain.ErrInvalidAvatarName
	}
	extension := strings.ToLower(filepath.Ext(dto.SrcFileName))
	formats := map[string]string{".jpg": "jpeg", ".jpeg": "jpeg", ".png": "png", ".gif": "gif"}
	expectedFormat, ok := formats[extension]
	if !ok {
		return domain.ErrInvalidAvatarName
	}
	config, format, err := image.DecodeConfig(dto.File)
	if err != nil || format != expectedFormat || config.Width <= 0 || config.Height <= 0 || int64(config.Width)*int64(config.Height) > 40_000_000 {
		return domain.ErrInvalidAvatarFormat
	}
	if _, err := dto.File.Seek(0, io.SeekStart); err != nil {
		return domain.ErrInvalidAvatarFormat
	}
	if _, format, err = image.Decode(dto.File); err != nil || format != expectedFormat {
		return domain.ErrInvalidAvatarFormat
	}
	if _, err := dto.File.Seek(0, io.SeekStart); err != nil {
		return domain.ErrInvalidAvatarFormat
	}
	contentType := "image/" + format
	key := strings.TrimLeft(dto.AvatarURL, "/"+u.avatarRepo.Bucket()+"/")
	if err := u.avatarRepo.Put(dto.Ctx, key, dto.File, dto.Size, contentType); err != nil {
		return err
	}
	return nil
}
