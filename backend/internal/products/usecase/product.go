package usecase

import "vibe_market/backend/internal/products/domain"

func (u *ProductUsecase) Product(id int64) (domain.ProductOutDTO, error) {
	var (
		result       = domain.ProductOutDTO{}
		err    error = nil
	)
	result.ProductByID, err = u.productRepo.SelectByID(id)
	if err != nil {
		return result, err
	}
	result.Price = round2(result.Price)
	result.PriceResult = round2((1 - float64(result.Discount)/100) * result.Price)

	return result, nil
}
