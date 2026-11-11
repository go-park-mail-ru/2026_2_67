package usecase

import (
	"context"
	"errors"
	"strings"
	"vibe_market/backend/internal/orders/domain"
)

type CreateOrderInDTO struct {
	BuyerID       int64
	PickupPointID int64
	Promocode     string
}

type CreateOrderOutDTO struct {
	OrderID int64 `json:"orderId"`
}

type OrderValidationErrors struct {
	PickupPointIDErrMessage string
	PromocodeErrMessage     string
}

func (v OrderValidationErrors) HasErrors() bool {
	return v.PickupPointIDErrMessage != "" || v.PromocodeErrMessage != ""
}

func CreateOrderUsecase(
	ctx context.Context,
	dto CreateOrderInDTO,
	repo domain.OrderRepo,
) (CreateOrderOutDTO, OrderValidationErrors, []domain.ProductStockError, error) {
	var valErrs OrderValidationErrors

	if dto.PickupPointID <= 0 {
		valErrs.PickupPointIDErrMessage = "pickupPointId должен быть положительным числом"
		return CreateOrderOutDTO{}, valErrs, nil, domain.ErrPickupPointInvalid
	}

	orderID, stockErrors, err := repo.CreateOrderTx(ctx, dto.BuyerID, dto.PickupPointID, strings.TrimSpace(dto.Promocode))
	if err != nil {
		if errors.Is(err, domain.ErrPickupPointInvalid) {
			valErrs.PickupPointIDErrMessage = "пункт выдачи не найден"
			return CreateOrderOutDTO{}, valErrs, nil, domain.ErrPickupPointInvalid
		}
		if errors.Is(err, domain.ErrPromocodeInvalid) {
			valErrs.PromocodeErrMessage = "промокод не найден или истек"
			return CreateOrderOutDTO{}, valErrs, nil, domain.ErrPromocodeInvalid
		}
		if len(stockErrors) > 0 {
			return CreateOrderOutDTO{}, valErrs, stockErrors, domain.ErrOrderStockInsufficient
		}
		return CreateOrderOutDTO{}, valErrs, nil, err
	}

	return CreateOrderOutDTO{OrderID: orderID}, valErrs, nil, nil
}
