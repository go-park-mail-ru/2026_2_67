package usecase

import (
	"context"

	"vibe_market/backend/internal/orders/domain"
)

const defaultOrdersLimit = 3
const maxOrdersLimit = 2 * defaultOrdersLimit

func (u *OrderUsecaseImpl) Orders(ctx context.Context, buyerID int64, params domain.OrdersParams) (domain.OrderPage, error) {
	if params.Limit == 0 {
		params.Limit = defaultOrdersLimit
	}
	if params.Limit > defaultOrdersLimit {
		params.Limit = maxOrdersLimit
	}
	if params.SortBy == "" {
		params.SortBy = domain.OrderSortByDescDate
	}
	return u.orderRepo.Orders(ctx, buyerID, params)
}
