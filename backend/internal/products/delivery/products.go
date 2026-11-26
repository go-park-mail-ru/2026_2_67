package delivery

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	"vibe_market/backend/internal/products/domain"
)

type cursorBase64 struct {
	CreatedAt time.Time `json:"created_at"`
	BuyCount  uint      `json:"buy_count"`
	Price     int       `json:"price"`
	ID        int64     `json:"id"`
	Rating    float64   `json:"rating"`
}

var (
	errPriceMinInvalid = errors.New("priceMin должен быть вещественным числом")
	errCursorBase64    = errors.New("cursor должен содержать created_at, buy_count, price, id, rating")
)

func encodeCursor(cursor domain.Cursor) (string, error) {
	b, err := json.Marshal(cursorBase64{CreatedAt: cursor.CreatedAt.UTC(),
		BuyCount: cursor.BuyCount, Price: cursor.Price, ID: cursor.ID, Rating: cursor.Rating})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodeCursor(cursor string) (cursorBase64, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return cursorBase64{}, errCursorBase64
	}
	var p cursorBase64
	if err := json.Unmarshal(raw, &p); err != nil {
		return cursorBase64{}, errCursorBase64
	}
	return p, nil
}

type productResponse struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	PictureURLs     []string `json:"pictureUrls"`
	Price           float64  `json:"price"`
	Discount        uint     `json:"discount"`
	PriceResult     float64  `json:"priceResult"`
	AvailableCount  uint     `json:"availableCount"`
	Rating          float64  `json:"rating"`
	ReviewsCount    uint     `json:"reviewsCount"`
	SellerID        int64    `json:"sellerId"`
	SellerName      string   `json:"sellerName"`
	SellerAvatarURL string   `json:"sellerAvatarURL"`
}

type productsResponse struct {
	Products       []productResponse `json:"products"`
	PriceResultMin float64           `json:"priceResultMin"`
	PriceResultMax float64           `json:"priceResultMax"`
	NextCursor     string            `json:"nextCursor"`
}

func (h *ProductsHandler) Products(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	var dto domain.ProductsSearchParams

	dto.Category = r.URL.Query().Get("category")
	dto.Seller = r.URL.Query()["seller"]
	dto.Query = r.URL.Query().Get("query")

	cursor := r.URL.Query().Get("cursor")
	if cursor != "" {
		decodedCursor, err := decodeCursor(cursor)
		if err != nil {
			logger.Warn(err.Error())
			dto.Cursor = nil
		} else {
			dto.Cursor = &domain.Cursor{ID: decodedCursor.ID,
				BuyCount: decodedCursor.BuyCount, Price: decodedCursor.Price,
				Rating: decodedCursor.Rating, CreatedAt: decodedCursor.CreatedAt}
		}
	}

	limitRaw := r.URL.Query().Get("limit")
	if limitRaw != "" {
		limit, err := strconv.ParseUint(limitRaw, 10, 32)
		if err != nil {
			logger.Warn(err.Error())
			dto.Limit = 0
		} else {
			dto.Limit = uint(limit)
		}
	}

	priceMinRaw := r.URL.Query().Get("priceMin")
	if priceMinRaw != "" {
		priceMin, err := strconv.ParseFloat(priceMinRaw, 64)
		if err != nil {
			logger.Warn(err.Error())
			dto.PriceMin = 0
		} else {
			dto.PriceMin = priceMin
		}
	}

	priceMaxRaw := r.URL.Query().Get("priceMax")
	if priceMaxRaw != "" {
		priceMax, err := strconv.ParseFloat(priceMaxRaw, 64)
		if err != nil {
			logger.Warn(err.Error())
			dto.PriceMax = nil
		} else {
			dto.PriceMax = &priceMax
		}
	}

	ratingMinRaw := r.URL.Query().Get("ratingMin")
	if ratingMinRaw != "" {
		ratingMin, err := strconv.ParseFloat(ratingMinRaw, 64)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			dto.RatingMin = ratingMin
		}
	}

	ratingMaxRaw := r.URL.Query().Get("ratingMax")
	if ratingMaxRaw != "" {
		ratingMax, err := strconv.ParseFloat(ratingMaxRaw, 64)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			dto.RatingMax = &ratingMax
		}
	}

	dto.SortBy = domain.SortByType(r.URL.Query().Get("sortBy"))

	page, err := h.productUsecase.Products(dto)

	if errors.Is(err, domain.ErrSortByInvalid) {
		logger.Warn(err.Error())
	} else if errors.Is(err, domain.ErrPriceRangeInvalid) {
		logger.Warn(err.Error())
	} else if errors.Is(err, domain.ErrRatingRangeInvalid) {
		logger.Warn(err.Error())
	} else if errors.Is(err, domain.ErrCursorInvalid) {
		logger.Warn(err.Error())
	} else if err != nil {
		logger.Error(err.Error())
		page.Products = []domain.Product{}
	}

	response := productsResponse{
		Products:       make([]productResponse, 0, len(page.Products)),
		PriceResultMin: float64(page.PriceResultMin) / 100,
		PriceResultMax: float64(page.PriceResultMax) / 100,
		NextCursor:     "",
	}
	if page.NextCursor != nil {
		cursor, err := encodeCursor(*page.NextCursor)
		if err != nil {
			logger.Error(err.Error())
		}
		response.NextCursor = cursor
	}

	for i, product := range page.Products {
		pictures := []string{}
		if page.PictureURLs[i] != nil {
			pictures = page.PictureURLs[i]
		}
		rating := page.Ratings[i]

		reviews := page.ReviewCounts[i]
		sellerAvatarURL := ""
		if page.Sellers[i].AvatarURL != nil {
			sellerAvatarURL = *page.Sellers[i].AvatarURL
		}
		response.Products = append(response.Products, productResponse{
			ID: product.ID, Name: product.Name, PictureURLs: pictures,
			Price: float64(product.Price) / 100, Discount: product.Discount,
			AvailableCount: product.AvailableCount,
			PriceResult:    float64(page.PricesResult[i]) / 100, Rating: rating, ReviewsCount: reviews,
			SellerID: page.Sellers[i].UserID, SellerName: page.Sellers[i].Name,
			SellerAvatarURL: sellerAvatarURL,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		logger.Error(err.Error())
	}
}
