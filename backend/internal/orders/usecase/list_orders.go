package usecase

import (
	"context"

	"vibe_market/backend/internal/orders/domain"
)

const defaultOrdersLimit = 15

func (u *OrderUsecaseImpl) Orders(ctx context.Context, buyerID int64, params domain.OrdersParams) (domain.OrderPage, error) {
	if params.Limit == 0 {
		params.Limit = defaultOrdersLimit
	}
	if params.SortBy == "" {
		params.SortBy = domain.OrderSortByDescDate
	}
	return u.orderRepo.Orders(ctx, buyerID, params)
}
