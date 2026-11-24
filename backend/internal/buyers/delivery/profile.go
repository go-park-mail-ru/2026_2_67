package delivery

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"mime"
	"net/http"
	"strconv"
	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/middleware"
	"vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/buyers/usecase"

	"go.uber.org/zap"
)

type buyerResponse struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic"`
	Telephone   string `json:"telephone"`
	Email       string `json:"email"`
	BirthDate   string `json:"birthDate"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatarUrl"`
}

func (h *BuyersHandler) Buyer(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	out, err := h.usecase.Buyer(userID)
	if errors.Is(err, authDomain.ErrUserIDNotFound) {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	err = encoder.Encode(buyerResponse{
		Name:        out.Name,
		Surname:     out.Surname,
		Patronymic:  out.Patronymic,
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

var errInvalidAvatar = errors.New("аватар должен быть корректным изображением JPEG, PNG или GIF с соответствующим расширением")

type patchBuyerResponse struct {
	NameErrMessage        string `json:"nameErrMessage"`
	SurnameErrMessage     string `json:"surnameErrMessage"`
	TelephoneErrMessage   string `json:"telephoneErrMessage"`
	EmailErrMessage       string `json:"emailErrMessage"`
	BirthDateErrMessage   string `json:"birthDateErrMessage"`
	DescriptionErrMessage string `json:"descriptionErrMessage"`
	AvatarErrMessage      string `json:"avatarErrMessage"`
}

func generatePictureName() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}

func getAvatarURL(avatarRepo domain.BuyerAvatarRepo, userID int64, fileName string) string {
	return fmt.Sprintf("/%s/buyers/%d/avatars/%s", avatarRepo.Bucket(), userID, fileName)
}

func writePatchError(w http.ResponseWriter, encoder *json.Encoder, out patchBuyerResponse) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = encoder.Encode(out)
}

func (h *BuyersHandler) PatchBuyer(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	accessToken, _ := r.Context().Value(middleware.AccessTokenPayloadKey).(*authDomain.AccessTokenPayload)
	if accessToken == nil || accessToken.UserID != userID {
		logger.Warn("В URL id не совпадает с accessToken",
			zap.Int64("PathID", userID))
		w.WriteHeader(http.StatusForbidden)
		return
	}

	encoder := json.NewEncoder(w)

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	mediaType, _, mediaTypeErr := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaTypeErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if mediaType == "multipart/form-data" {
		err = r.ParseMultipartForm(maxAvatarSize)
	} else {
		logger.Warn("Неправильная форма")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		logger.Warn(err.Error())
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writePatchError(w, encoder, patchBuyerResponse{AvatarErrMessage: fmt.Sprintf("Слишком большой файл. Допустимый размер: %d", maxAvatarSize)})
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	// получение файла
	file, header, err := r.FormFile("avatar")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) && mediaType == "multipart/form-data" {
			writePatchError(w, encoder, patchBuyerResponse{AvatarErrMessage: "Некорректная аватарка"})
			return
		}
		header = nil
	}
	var avatarURL *string
	if file != nil {
		defer file.Close()
		if header.Size > maxAvatarSize {
			writePatchError(w, encoder, patchBuyerResponse{AvatarErrMessage: fmt.Sprintf("Слишком большой файл. Допустимый размер: %d", maxAvatarSize)})
			return
		}
		fileName := generatePictureName()
		if fileName == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		value := getAvatarURL(h.avatarRepo, userID, fileName)
		avatarURL = &value
		err = h.usecase.PatchBuyerAvatar(usecase.PatchBuyerAvatarInDTO{Ctx: r.Context(), SrcFileName: header.Filename,
			File: file, AvatarURL: value, Size: header.Size})
		if errors.Is(err, domain.ErrAvatarNameInvalid) || errors.Is(err, domain.ErrAvatarFormatInvalid) {
			logger.Warn(err.Error())
			writePatchError(w, encoder, patchBuyerResponse{AvatarErrMessage: err.Error()})
			return
		}
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// установление новых параметров
	out, err := h.usecase.PatchBuyer(usecase.PatchBuyerInDTO{
		UserID: userID, Name: r.FormValue("name"), Surname: r.FormValue("surname"),
		Telephone: r.FormValue("telephone"), Email: r.FormValue("email"),
		BirthDate: r.FormValue("birthDate"), Description: r.FormValue("description"), AvatarURL: avatarURL,
	})
	if err != nil {
		if errors.Is(err, authDomain.ErrUserIDNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, authDomain.ErrEmailAlreadyExists) {
			writePatchError(w, encoder, patchBuyerResponse{EmailErrMessage: err.Error()})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := encoder.Encode(buyerResponse{Name: out.Name, Surname: out.Surname,
		Telephone: out.Telephone, Email: out.Email, BirthDate: out.BirthDate,
		Description: out.Description, AvatarURL: out.AvatarURL}); err != nil {
		logger.Error(err.Error())
	}
}
