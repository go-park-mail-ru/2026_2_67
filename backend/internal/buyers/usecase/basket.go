package usecase

import (
	"errors"
	"strings"
	"time"
	"vibe_market/backend/internal/buyers/domain"
)

type BasketProductItemDTO struct {
	ID          int64   `json:"id"`
	Category    string  `json:"category"`
	Name        string  `json:"name"`
	PictureURL  string  `json:"pictureUrl"`
	Price       float64 `json:"price"`
	Discount    int     `json:"discount"`
	PriceResult float64 `json:"priceResult"`
	Count       int     `json:"count"`
}

type GetBasketOutDTO struct {
	BasketProducts  []BasketProductItemDTO `json:"basketProducts"`
	BasketSum       float64                `json:"basketSum"`
	BasketSumResult float64                `json:"basketSumResult"`
}

type PreviewInDTO struct {
	Longitude float64
	Latitude  float64
	Promocode string
}

type PreviewProductDTO struct {
	Name               string  `json:"name"`
	Price              float64 `json:"price"`
	PriceWithPromocode float64 `json:"priceWithPromocode"`
	PriceWithDiscount  float64 `json:"priceWithDiscount"`
	PriceResult        float64 `json:"priceResult"`
}

type PickupPointCoordDTO struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type PreviewOutDTO struct {
	Products             []PreviewProductDTO `json:"products"`
	ExpectedDeliveryDate string              `json:"expectedDeliveryDate"`
	PickupPoint          PickupPointCoordDTO `json:"pickupPoint"`
	Price                float64             `json:"price"`
	PriceWithPromocode   float64             `json:"priceWithPromocode"`
	PriceWithDiscount    float64             `json:"priceWithDiscount"`
	PriceResult          float64             `json:"priceResult"`
}

type PreviewValidationErrors struct {
	LongitudeErrMessage string
	LatitudeErrMessage  string
	PromocodeErrMessage string
}

func (v PreviewValidationErrors) HasErrors() bool {
	return v.LongitudeErrMessage != "" || v.LatitudeErrMessage != "" || v.PromocodeErrMessage != ""
}

func GetBasketUsecase(buyerID int64, repo domain.BasketRepo) (GetBasketOutDTO, error) {
	summary, err := repo.GetBasket(buyerID)
	if err != nil {
		return GetBasketOutDTO{}, err
	}

	items := make([]BasketProductItemDTO, 0, len(summary.Items))
	for _, it := range summary.Items {
		items = append(items, BasketProductItemDTO{
			ID:          it.ProductID,
			Category:    it.CategoryName,
			Name:        it.Name,
			PictureURL:  it.PictureURL,
			Price:       it.Price,
			Discount:    it.Discount,
			PriceResult: it.PriceResult,
			Count:       it.Count,
		})
	}

	return GetBasketOutDTO{
		BasketProducts:  items,
		BasketSum:       summary.BasketSum,
		BasketSumResult: summary.BasketSumResult,
	}, nil
}

func AddProductUsecase(buyerID int64, productID int64, count int, repo domain.BasketRepo) error {
	if count <= 0 {
		return domain.ErrCountInvalid
	}
	return repo.AddProduct(buyerID, productID, count)
}

func UpdateProductCountUsecase(buyerID int64, productID int64, count int, repo domain.BasketRepo) error {
	if count < 0 {
		return domain.ErrCountInvalid
	}
	return repo.UpdateProductCount(buyerID, productID, count)
}

func PreviewBasketUsecase(
	buyerID int64,
	dto PreviewInDTO,
	repo domain.BasketRepo,
) (PreviewOutDTO, PreviewValidationErrors, []domain.ProductStockError, error) {
	var valErrs PreviewValidationErrors

	if dto.Longitude < -180 || dto.Longitude > 180 {
		valErrs.LongitudeErrMessage = "longitude должен быть в диапазоне от -180 до 180"
	}
	if dto.Latitude < -90 || dto.Latitude > 90 {
		valErrs.LatitudeErrMessage = "latitude должен быть в диапазоне от -90 до 90"
	}

	var promo *domain.Promocode
	promoCode := strings.TrimSpace(dto.Promocode)
	if promoCode != "" {
		p, err := repo.GetPromocode(promoCode)
		if err != nil {
			if errors.Is(err, domain.ErrPromocodeNotFoundOrExpired) {
				valErrs.PromocodeErrMessage = "промокод не найден или истек"
			} else {
				return PreviewOutDTO{}, valErrs, nil, err
			}
		} else {
			promo = p
		}
	}

	if valErrs.HasErrors() {
		return PreviewOutDTO{}, valErrs, nil, domain.ErrInvalidCoordinates
	}

	summary, err := repo.GetBasket(buyerID)
	if err != nil {
		return PreviewOutDTO{}, valErrs, nil, err
	}

	if len(summary.Items) == 0 {
		return PreviewOutDTO{}, valErrs, nil, domain.ErrBasketEmpty
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
		return PreviewOutDTO{}, valErrs, stockErrors, domain.ErrCountExceeded
	}

	pickup, err := repo.GetNearestPickupPoint(dto.Latitude, dto.Longitude)
	if err != nil {
		return PreviewOutDTO{}, valErrs, nil, err
	}

	var totalOriginal float64
	var totalWithPromocode float64
	var totalWithDiscount float64
	var totalResult float64

	products := make([]PreviewProductDTO, 0, len(summary.Items))
	for _, it := range summary.Items {
		pOriginal := it.Price
		pWithDiscount := it.Price * (1.0 - float64(it.Discount)/100.0)

		pWithPromo := it.Price
		discountFactor := 1.0 - float64(it.Discount)/100.0

		if promo != nil && (promo.AvailableCategoryID == nil || *promo.AvailableCategoryID == it.CategoryID) {
			pWithPromo = it.Price * (1.0 - float64(promo.Discount)/100.0)
			discountFactor *= 1.0 - float64(promo.Discount)/100.0
		}

		pResult := it.Price * discountFactor

		totalOriginal += pOriginal * float64(it.Count)
		totalWithDiscount += pWithDiscount * float64(it.Count)
		totalWithPromocode += pWithPromo * float64(it.Count)
		totalResult += pResult * float64(it.Count)

		products = append(products, PreviewProductDTO{
			Name:               it.Name,
			Price:              pOriginal,
			PriceWithPromocode: pWithPromo,
			PriceWithDiscount:  pWithDiscount,
			PriceResult:        pResult,
		})
	}

	deliveryDate := time.Now().AddDate(0, 0, 3).Format("2006-01-02")

	out := PreviewOutDTO{
		Products:             products,
		ExpectedDeliveryDate: deliveryDate,
		PickupPoint: PickupPointCoordDTO{
			Longitude: pickup.Longitude,
			Latitude:  pickup.Latitude,
		},
		Price:              totalOriginal,
		PriceWithPromocode: totalWithPromocode,
		PriceWithDiscount:  totalWithDiscount,
		PriceResult:        totalResult,
	}

	return out, valErrs, nil, nil
}
