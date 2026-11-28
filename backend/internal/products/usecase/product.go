package usecase

import (
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/products/domain"
)

type ProductOutDTO struct {
	domain.ProductByID
	PriceResult int
}

func (u *ProductUsecaseImpl) Product(id int64, buyerID *int64) (ProductOutDTO, error) {
	var (
		result            = ProductOutDTO{}
		err         error = nil
		repoBuyerID int64 = 0
	)
	if buyerID != nil {
		repoBuyerID = *buyerID
	}
	result.ProductByID, err = u.productRepo.SelectByID(id, repoBuyerID)
	if err != nil {
		return result, err
	}
	result.PriceResult = common.CalcPriceWithDiscount(result.Price, result.Discount)
	return result, nil
}
