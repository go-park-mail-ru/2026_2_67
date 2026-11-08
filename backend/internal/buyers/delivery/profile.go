package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/buyers/usecase"

	"go.uber.org/zap"
)

type getBuyerProfileResponse struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Telephone   string `json:"telephone"`
	Email       string `json:"email"`
	BirthDate   string `json:"birthDate"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatarUrl"`
}

func (h *BuyersHandler) GetBuyerProfile(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	out, err := usecase.GetBuyerProfileUsecase(userID, h.buyerRepo)
	if errors.Is(err, domain.ErrUserIDNotFound) {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if !errors.Is(err, nil) {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(getBuyerProfileResponse{
		Name:        out.Name,
		Surname:     out.Surname,
		Telephone:   out.Telephone,
		Email:       out.Email,
		BirthDate:   out.BirthDate,
		Description: out.Description,
		AvatarURL:   out.AvatarURL,
	})
	if err != nil {
		logger.Error(err.Error())
	}
}

const maxAvatarSize = 16 << 20
const maxBodySize = maxAvatarSize + 1<<20

type patchBuyerProfileErr struct {
	NameErrMessage        string `json:"nameErrMessage"`
	SurnameErrMessage     string `json:"surnameErrMessage"`
	TelephoneErrMessage   string `json:"telephoneErrMessage"`
	EmailErrMessage       string `json:"emailErrMessage"`
	BirthDateErrMessage   string `json:"birthDateErrMessage"`
	DescriptionErrMessage string `json:"descriptionErrMessage"`
	AvatarErrMessage      string `json:"avatarErrMessage"`
}

func (h *BuyersHandler) PatchBuyerProfile(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// проверка авторизации
	accessToken, err := middleware.ParseAccessTokenPayload(r, h.jwtSecret)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if accessToken.UserID != userID {
		logger.Warn("В URL id не совпадает с accessToken",
			zap.Int64("PathID", userID), zap.Int64("UserID", accessToken.UserID))
		w.WriteHeader(http.StatusForbidden)
		return
	}

	encoder := json.NewEncoder(w)

	// проверка размера файла
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	err = r.ParseMultipartForm(maxAvatarSize)
	if err != nil {
		logger.Warn(err.Error())
		var maxErr *http.MaxBytesError
		if !errors.As(err, &maxErr) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusUnprocessableEntity)
		out := patchBuyerProfileErr{
			AvatarErrMessage: fmt.Sprintf("Слишком большой файл. Допустимый размер: %d", maxAvatarSize),
		}
		err = encoder.Encode(out)
		if err != nil {
			logger.Error(err.Error())
		}
		return
	}
	defer r.MultipartForm.RemoveAll()

	// получение файла
	file, header, err := r.FormFile("avatar")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			writePatchError(w, encoder, patchBuyerProfileErr{AvatarErrMessage: "Некорректная аватарка"})
			return
		}
		header = nil
	}
	if file != nil {
		defer file.Close()
		if header.Size > maxAvatarSize {
			writePatchError(w, encoder, patchBuyerProfileErr{AvatarErrMessage: fmt.Sprintf("Слишком большой файл. Допустимый размер: %d", maxAvatarSize)})
			return
		}
	}

	// загрузка аватарки
	avatarURL := (*string)(nil)
	if file != nil {
		avatarURL, err = h.uploadAvatar(r.Context(), userID, header.Filename, file, header.Size,
			header.Header.Get("Content-Type"))
	}
	if err != nil {
		logger.Error(err.Error())
		writePatchError(w, encoder, patchBuyerProfileErr{AvatarErrMessage: "Не удалось загрузить аватарку"})
		return
	}
	// установление новых параметров
	out, err := usecase.PatchBuyerProfile(usecase.PatchBuyerProfileInDTO{
		UserID: userID, Name: r.FormValue("name"), Surname: r.FormValue("surname"),
		Telephone: r.FormValue("telephone"), Email: r.FormValue("email"),
		BirthDate: r.FormValue("birthDate"), Description: r.FormValue("description"), AvatarURL: avatarURL,
	}, h.buyerRepo, h.userRepo)
	if err != nil {
		if errors.Is(err, domain.ErrUserIDNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		writePatchError(w, encoder, patchBuyerProfileErr{DescriptionErrMessage: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := encoder.Encode(getBuyerProfileResponse{Name: out.Name, Surname: out.Surname,
		Telephone: out.Telephone, Email: out.Email, BirthDate: out.BirthDate,
		Description: out.Description, AvatarURL: out.AvatarURL}); err != nil {
		logger.Error(err.Error())
	}
}

func (h *BuyersHandler) uploadAvatar(ctx context.Context, userID int64, fileName string, body io.Reader,
	size int64, contentType string) (*string, error) {
	fileName = filepath.Base(fileName)
	if fileName == "." || fileName == "" || strings.Contains(fileName, "\\") {
		return nil, errors.New("invalid filename")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	objectName := fmt.Sprintf("buyers/%d/%s", userID, fileName)
	if err := h.avatarStorage.Put(ctx, objectName, body, size, contentType); err != nil {
		return nil, err
	}
	return &objectName, nil
}

func writePatchError(w http.ResponseWriter, encoder *json.Encoder, out patchBuyerProfileErr) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = encoder.Encode(out)
}
