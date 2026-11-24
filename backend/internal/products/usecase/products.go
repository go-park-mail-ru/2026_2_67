package usecase

import (
	"errors"
	"vibe_market/backend/internal/common"
	"vibe_market/backend/internal/products/domain"
	sellerDomain "vibe_market/backend/internal/sellers/domain"
)

type ProductsOutDTO struct {
	Products       []domain.Product
	Sellers        []sellerDomain.Seller
	PricesResult   []int
	PictureURLs    [][]string
	Ratings        []float64
	ReviewCounts   []uint
	PriceResultMin int
	PriceResultMax int
	NextCursor     *domain.Cursor
}

const productsDefaultLimit = 20
const productsMaxLimit = 2 * productsDefaultLimit

func (u *ProductUsecaseImpl) Products(dto domain.ProductsSearchParams) (ProductsOutDTO, error) {
	if dto.SortBy == "" {
		dto.SortBy = domain.SortByNew
	}
	if dto.Limit > productsMaxLimit {
		dto.Limit = productsMaxLimit
	}
	if dto.Limit == 0 {
		dto.Limit = productsDefaultLimit
	}
	products, err := u.productRepo.Search(dto)
	if errors.Is(err, domain.ErrSortByInvalid) {
		return ProductsOutDTO{}, err
	}
	if errors.Is(err, domain.ErrPriceRangeInvalid) {
		return ProductsOutDTO{}, err
	}
	if errors.Is(err, domain.ErrCursorInvalid) {
		return ProductsOutDTO{}, err
	}
	if err != nil {
		return ProductsOutDTO{}, err
	}

	size := len(products.Products)
	pricesResult := make([]int, size)
	for i := range size {
		pricesResult[i] = common.CalcPriceWithDiscount(products.Products[i].Price, products.Products[i].Discount)
	}

	return ProductsOutDTO{
		Products: products.Products, Sellers: products.Sellers,
		PricesResult: pricesResult, PictureURLs: products.PictureURLs,
		Ratings: products.Ratings, ReviewCounts: products.ReviewCounts,
		PriceResultMin: products.PriceResultMin, PriceResultMax: products.PriceResultMax,
		NextCursor: products.NextCursor,
	}, nil
}
