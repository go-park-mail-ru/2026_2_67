package usecase

import (
	"errors"
	"math"
	"vibe_market/backend/internal/products/domain"
)

func round2(n float64) float64 {
	return math.RoundToEven(n*100) / 100
}

func (u *ProductUsecase) Products(dto domain.ProductsSearchParams) (domain.ProductsOutDTO, error) {
	products, err := u.productRepo.Search(dto)
	if errors.Is(err, domain.ErrInvalidSortBy) {
		return domain.ProductsOutDTO{}, err
	}
	if errors.Is(err, domain.ErrInvalidPriceRange) {
		return domain.ProductsOutDTO{}, err
	}
	if errors.Is(err, domain.ErrInvalidCursor) {
		return domain.ProductsOutDTO{}, err
	}
	if err != nil {
		return domain.ProductsOutDTO{}, err
	}

	size := len(products.Products)
	pricesResult := make([]float64, size)
	for i := range size {
		pricesResult[i] = (1 - float64(products.Products[i].Discount)/100) * products.Products[i].Price
		pricesResult[i] = round2(pricesResult[i])
	}

	return domain.ProductsOutDTO{
		Products: products.Products, Sellers: products.Sellers,
		PricesResult: pricesResult, PictureURLs: products.PictureURLs,
		Ratings: products.Ratings, ReviewCounts: products.ReviewCounts, NextCursor: products.NextCursor,
	}, nil
}
