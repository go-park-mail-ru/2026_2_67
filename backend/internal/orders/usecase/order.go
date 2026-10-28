package usecase

import (
	"context"
	"errors"
	"strings"
	"vibe_market/backend/internal/orders/domain"
)

type CreateOrderInDTO struct {
	BuyerID   int64
	Longitude float64
	Latitude  float64
	Promocode string
}

type CreateOrderOutDTO struct {
	OrderID int64 `json:"orderId"`
}

type OrderValidationErrors struct {
	LongitudeErrMessage string
	LatitudeErrMessage  string
	PromocodeErrMessage string
}

func (v OrderValidationErrors) HasErrors() bool {
	return v.LongitudeErrMessage != "" || v.LatitudeErrMessage != "" || v.PromocodeErrMessage != ""
}

func CreateOrderUsecase(
	ctx context.Context,
	dto CreateOrderInDTO,
	repo domain.OrderRepo,
) (CreateOrderOutDTO, OrderValidationErrors, []domain.ProductStockError, error) {
	var valErrs OrderValidationErrors

	if dto.Longitude < -180 || dto.Longitude > 180 {
		valErrs.LongitudeErrMessage = "longitude должен быть в диапазоне от -180 до 180"
	}
	if dto.Latitude < -90 || dto.Latitude > 90 {
		valErrs.LatitudeErrMessage = "latitude должен быть в диапазоне от -90 до 90"
	}

	if valErrs.HasErrors() {
		return CreateOrderOutDTO{}, valErrs, nil, domain.ErrInvalidCoordinates
	}

	orderID, stockErrors, err := repo.CreateOrderTx(ctx, dto.BuyerID, dto.Latitude, dto.Longitude, strings.TrimSpace(dto.Promocode))
	if err != nil {
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
