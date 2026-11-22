package usecase

import (
	"errors"
	"strings"
	"time"
	"vibe_market/backend/internal/buyers/domain"
	"vibe_market/backend/internal/common"
)

type BasketProductItemOutDTO struct {
	ID          int64   `json:"id"`
	Category    string  `json:"category"`
	Name        string  `json:"name"`
	PictureURL  string  `json:"pictureUrl"`
	Price       float64 `json:"price"`
	Discount    int     `json:"discount"`
	PriceResult float64 `json:"priceResult"`
	Count       int     `json:"count"`
}

type BasketOutDTO struct {
	BasketProducts  []BasketProductItemOutDTO `json:"basketProducts"`
	BasketSum       float64                   `json:"basketSum"`
	BasketSumResult float64                   `json:"basketSumResult"`
}

type IsValidOutDTO struct {
	Products           []int64 `json:"products"`
	NewAvailableCounts []uint
}

type PreviewBasketInDTO struct {
	Longitude float64
	Latitude  float64
	Promocode string
}

type PreviewBasketProductOutDTO struct {
	Name               string  `json:"name"`
	Price              float64 `json:"price"`
	Discount           int     `json:"discount"`
	PriceWithPromocode float64 `json:"priceWithPromocode"`
	PriceWithDiscount  float64 `json:"priceWithDiscount"`
	PriceResult        float64 `json:"priceResult"`
}

type PreviewBasketPickupPointOutDTO struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type PreviewBasketOutDTO struct {
	Products             []PreviewBasketProductOutDTO   `json:"products"`
	ExpectedDeliveryDate string                         `json:"expectedDeliveryDate"`
	PickupPoint          PreviewBasketPickupPointOutDTO `json:"pickupPoint"`
	Price                float64                        `json:"price"`
	PriceWithPromocode   float64                        `json:"priceWithPromocode"`
	PriceWithDiscount    float64                        `json:"priceWithDiscount"`
	PriceResult          float64                        `json:"priceResult"`
}

type PreviewBasketValidationErrors struct {
	LongitudeErrMessage string
	LatitudeErrMessage  string
	PromocodeErrMessage string
}

func (v PreviewBasketValidationErrors) HasErrors() bool {
	return v.LongitudeErrMessage != "" || v.LatitudeErrMessage != "" || v.PromocodeErrMessage != ""
}

func (u *BasketUsecaseImpl) Basket(buyerID int64) (BasketOutDTO, error) {
	summary, err := u.basketRepo.GetBasket(buyerID)
	if err != nil {
		return BasketOutDTO{}, err
	}

	items := make([]BasketProductItemOutDTO, 0, len(summary.Items))
	for _, it := range summary.Items {
		items = append(items, BasketProductItemOutDTO{
			ID:          it.ProductID,
			Category:    it.CategoryName,
			Name:        it.Name,
			PictureURL:  it.PictureURL,
			Price:       float64(it.Price) / 100,
			Discount:    it.Discount,
			PriceResult: float64(it.PriceResult) / 100,
			Count:       it.Count,
		})
	}

	return BasketOutDTO{
		BasketProducts:  items,
		BasketSum:       float64(summary.BasketSum) / 100,
		BasketSumResult: float64(summary.BasketSumResult) / 100,
	}, nil
}

func (u *BasketUsecaseImpl) IsValid(buyerID int64) (IsValidOutDTO, error) {
	summary, err := u.basketRepo.GetBasket(buyerID)
	if err != nil {
		return IsValidOutDTO{}, err
	}

	products := make([]int64, 0)
	newAvailableCounts := make([]uint, 0)
	for _, item := range summary.Items {
		if item.Count > item.AvailableCount {
			products = append(products, item.ProductID)
			newAvailableCounts = append(newAvailableCounts, uint(item.AvailableCount))
		}
	}

	return IsValidOutDTO{Products: products, NewAvailableCounts: newAvailableCounts}, nil
}

func (u *BasketUsecaseImpl) AddProduct(buyerID int64, productID int64, count int) error {
	if count <= 0 {
		return domain.ErrCountInvalid
	}
	return u.basketRepo.AddProduct(buyerID, productID, count)
}

func (u *BasketUsecaseImpl) UpdateProductCount(buyerID int64, productID int64, count int) error {
	if count < 0 {
		return domain.ErrCountInvalid
	}
	return u.basketRepo.UpdateProductCount(buyerID, productID, count)
}

func (u *BasketUsecaseImpl) PreviewBasket(
	buyerID int64,
	dto PreviewBasketInDTO,
) (PreviewBasketOutDTO, PreviewBasketValidationErrors, []domain.ProductStockError, error) {
	var valErrs PreviewBasketValidationErrors

	if dto.Longitude < -180 || dto.Longitude > 180 {
		valErrs.LongitudeErrMessage = "longitude должен быть в диапазоне от -180 до 180"
	}
	if dto.Latitude < -90 || dto.Latitude > 90 {
		valErrs.LatitudeErrMessage = "latitude должен быть в диапазоне от -90 до 90"
	}

	var promo *domain.Promocode
	promoCode := strings.TrimSpace(dto.Promocode)
	if promoCode != "" {
		p, err := u.basketRepo.GetPromocode(promoCode)
		if err != nil {
			if errors.Is(err, domain.ErrPromocodeNotFoundOrExpired) {
				valErrs.PromocodeErrMessage = "промокод не найден или истек"
			} else {
				return PreviewBasketOutDTO{}, valErrs, nil, err
			}
		} else {
			promo = p
		}
	}

	if valErrs.HasErrors() {
		return PreviewBasketOutDTO{}, valErrs, nil, domain.ErrInvalidCoordinates
	}

	summary, err := u.basketRepo.GetBasket(buyerID)
	if err != nil {
		return PreviewBasketOutDTO{}, valErrs, nil, err
	}

	if len(summary.Items) == 0 {
		return PreviewBasketOutDTO{}, valErrs, nil, domain.ErrBasketEmpty
	}

	stockErrors := make([]domain.ProductStockError, 0)
	for _, it := range summary.Items {
		if it.Count > it.AvailableCount {
			stockErrors = append(stockErrors, domain.ProductStockError{
				ProductID:  it.ProductID,
				Name:       it.Name,
				Seller:     it.SellerName,
				ErrMessage: "товары закончились или их недостаточно на складе",
			})
		}
	}

	if len(stockErrors) > 0 {
		return PreviewBasketOutDTO{}, valErrs, stockErrors, domain.ErrCountExceeded
	}

	pickup, err := u.basketRepo.GetNearestPickupPoint(dto.Latitude, dto.Longitude)
	if err != nil {
		return PreviewBasketOutDTO{}, valErrs, nil, err
	}

	var totalOriginal int
	var totalWithPromocode int
	var totalWithDiscount int
	var totalResult int

	products := make([]PreviewBasketProductOutDTO, 0, len(summary.Items))
	for _, it := range summary.Items {
		pOriginal := it.Price
		pWithDiscount := common.CalcPriceWithDiscount(it.Price, uint(it.Discount))

		pWithPromo := it.Price
		pResult := pWithDiscount

		if promo != nil && (promo.AvailableCategoryID == nil || *promo.AvailableCategoryID == it.CategoryID) {
			pWithPromo = common.CalcPriceWithDiscount(it.Price, uint(promo.Discount))
			pResult = common.CalcPriceWithDiscount(pWithDiscount, uint(promo.Discount))
		}

		totalOriginal += pOriginal * it.Count
		totalWithDiscount += pWithDiscount * it.Count
		totalWithPromocode += pWithPromo * it.Count
		totalResult += pResult * it.Count

		products = append(products, PreviewBasketProductOutDTO{
			Name:               it.Name,
			Price:              float64(pOriginal) / 100,
			Discount:           it.Discount,
			PriceWithPromocode: float64(pWithPromo) / 100,
			PriceWithDiscount:  float64(pWithDiscount) / 100,
			PriceResult:        float64(pResult) / 100,
		})
	}

	deliveryDate := time.Now().AddDate(0, 0, 3).Format("2006-01-02")

	out := PreviewBasketOutDTO{
		Products:             products,
		ExpectedDeliveryDate: deliveryDate,
		PickupPoint: PreviewBasketPickupPointOutDTO{
			Longitude: pickup.Longitude,
			Latitude:  pickup.Latitude,
		},
		Price:              float64(totalOriginal) / 100,
		PriceWithPromocode: float64(totalWithPromocode) / 100,
		PriceWithDiscount:  float64(totalWithDiscount) / 100,
		PriceResult:        float64(totalResult) / 100,
	}

	return out, valErrs, nil, nil
}
