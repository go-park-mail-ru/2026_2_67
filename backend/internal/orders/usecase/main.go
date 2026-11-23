package usecase

import (
	"context"
	"vibe_market/backend/internal/orders/domain"
)

type OrderUsecase interface {
	Orders(ctx context.Context, buyerID int64, params domain.OrdersParams) (domain.OrderPage, error)
}

type OrderUsecaseImpl struct {
	orderRepo domain.OrderRepo
}

func NewOrderUsecase(orderRepo domain.OrderRepo) *OrderUsecaseImpl {
	return &OrderUsecaseImpl{orderRepo: orderRepo}
}
