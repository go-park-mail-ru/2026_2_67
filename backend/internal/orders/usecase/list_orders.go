package usecase

import (
	"context"

	"vibe_market/backend/internal/orders/domain"
)

func ListOrders(ctx context.Context, buyerID int64, repo domain.OrderRepo) ([]domain.OrderDetails, error) {
	return repo.ListByBuyer(ctx, buyerID)
}
