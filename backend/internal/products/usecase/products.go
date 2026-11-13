package usecase

import (
	"errors"
	"math"
	"vibe_market/backend/internal/products/domain"
	sellerDomain "vibe_market/backend/internal/sellers/domain"
)

type ProductsOutDTO struct {
	Products     []domain.Product
	Sellers      []sellerDomain.Seller
	PricesResult []float64
	PictureURLs  [][]string
	Ratings      []float64
	ReviewCounts []uint
	NextCursor   *domain.Cursor
}

func round2(n float64) float64 {
	return math.RoundToEven(n*100) / 100
}

func Products(dto domain.ProductsSearchParams, productsRepo domain.ProductRepo) (ProductsOutDTO, error) {
	products, err := productsRepo.Search(dto)
	if errors.Is(err, domain.ErrInvalidSortBy) {
		return ProductsOutDTO{}, err
	}
	if errors.Is(err, domain.ErrInvalidPriceRange) {
		return ProductsOutDTO{}, err
	}
	if errors.Is(err, domain.ErrInvalidCursor) {
		return ProductsOutDTO{}, err
	}
	if err != nil {
		return ProductsOutDTO{}, err
	}

	size := len(products.Products)
	pricesResult := make([]float64, size)
	for i := range size {
		pricesResult[i] = (1 - float64(products.Products[i].Discount)/100) * products.Products[i].Price
		pricesResult[i] = round2(pricesResult[i])
	}

	return ProductsOutDTO{
		Products: products.Products, Sellers: products.Sellers,
		PricesResult: pricesResult, PictureURLs: products.PictureURLs,
		Ratings: products.Ratings, ReviewCounts: products.ReviewCounts, NextCursor: products.NextCursor,
	}, nil
}
