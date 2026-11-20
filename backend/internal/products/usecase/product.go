package usecase

import (
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/products/domain"
)

func (u *ProductUsecase) Product(id int64) (domain.ProductOutDTO, error) {
	var (
		result       = domain.ProductOutDTO{}
		err    error = nil
	)
	result.ProductByID, err = u.productRepo.SelectByID(id)
	if err != nil {
		return result, err
	}
	result.PriceResult = common.CalcPriceWithDiscount(result.Price, result.Discount)
	return result, nil
}
