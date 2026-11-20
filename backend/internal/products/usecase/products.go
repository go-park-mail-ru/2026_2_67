package usecase

import (
	"errors"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/products/domain"
)

func (u *ProductUsecase) Products(dto domain.ProductsSearchParams) (domain.ProductsOutDTO, error) {
	products, err := u.productRepo.Search(dto)
	if errors.Is(err, domain.ErrSortByInvalid) {
		return domain.ProductsOutDTO{}, err
	}
	if errors.Is(err, domain.ErrPriceRangeInvalid) {
		return domain.ProductsOutDTO{}, err
	}
	if errors.Is(err, domain.ErrCursorInvalid) {
		return domain.ProductsOutDTO{}, err
	}
	if err != nil {
		return domain.ProductsOutDTO{}, err
	}

	size := len(products.Products)
	pricesResult := make([]int, size)
	for i := range size {
		pricesResult[i] = common.CalcPriceWithDiscount(products.Products[i].Price, products.Products[i].Discount)
	}

	return domain.ProductsOutDTO{
		Products: products.Products, Sellers: products.Sellers,
		PricesResult: pricesResult, PictureURLs: products.PictureURLs,
		Ratings: products.Ratings, ReviewCounts: products.ReviewCounts, NextCursor: products.NextCursor,
	}, nil
}
