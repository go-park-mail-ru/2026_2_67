package controller

import "vibe_market/backend/internal/buyers/domain"

type BuyersHandler struct {
	buyerRepo domain.BuyerRepo
}

func NewBuyersHandler(buyerRepo domain.BuyerRepo) *BuyersHandler {
	return &BuyersHandler{buyerRepo}
}
