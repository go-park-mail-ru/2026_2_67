package usecase

import "vibe_market/backend/internal/products/domain"

type ProductOutDTO struct {
	domain.ProductByID
	PriceResult float64
}

func Product(id int64, productRepo domain.ProductRepo) (ProductOutDTO, error) {
	var (
		result       = ProductOutDTO{}
		err    error = nil
	)
	result.ProductByID, err = productRepo.SelectByID(id)
	if err != nil {
		return result, err
	}
	result.Price = round2(result.Price)
	result.PriceResult = round2((1 - float64(result.Discount)/100) * result.Price)

	return result, nil
}
