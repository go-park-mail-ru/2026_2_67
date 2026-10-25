package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/buyers/usecase"
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
