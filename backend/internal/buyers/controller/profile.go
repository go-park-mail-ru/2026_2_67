package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"vibe_market/backend/internal/buyers/usecase"
)

type getBuyerProfileResponse struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Telephone   string `json:"phoneNumber"`
	Email       string `json:"email"`
	BirthDate   string `json:"birthDate"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatarUrl"`
}

func (h *BuyersHandler) GetBuyerProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	out, err := usecase.GetBuyerProfileUsecase(userID, h.buyerRepo)
	if errors.Is(err, usecase.ErrBuyerNotFound) {
		w.WriteHeader(http.StatusNotFound)
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
		fmt.Println(err)
		return
	}
}
