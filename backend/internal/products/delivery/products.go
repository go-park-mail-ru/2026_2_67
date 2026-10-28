package delivery

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	"vibe_market/backend/internal/products/domain"
	"vibe_market/backend/internal/products/usecase"
)

type cursorBase64 struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int64     `json:"id"`
	Rating    float64   `json:"rating"`
}

var (
	ErrInvalidPriceMin     = errors.New("priceMin должен быть вещественным")
	ErrInvalidCursorBase64 = errors.New("cursor должен содержать created_at, id, rating")
)

func encodeCursor(t time.Time, id int64, rating float64) (string, error) {
	b, err := json.Marshal(cursorBase64{CreatedAt: t.UTC(), ID: id, Rating: rating})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodeCursor(cursor string) (cursorBase64, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return cursorBase64{}, ErrInvalidCursorBase64
	}
	var p cursorBase64
	if err := json.Unmarshal(raw, &p); err != nil {
		return cursorBase64{}, ErrInvalidCursorBase64
	}
	if p.CreatedAt.IsZero() {
		return cursorBase64{}, ErrInvalidCursorBase64
	}
	return p, nil
}

type productResponse struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	PictureURLs  []string `json:"pictureUrls"`
	Price        float64  `json:"price"`
	Discount     int      `json:"discount"`
	PriceResult  float64  `json:"priceResult"`
	Rating       float64  `json:"rating"`
	ReviewsCount uint     `json:"reviewsCount"`
}

type productsResponse struct {
	Products   []productResponse `json:"products"`
	NextCursor string            `json:"nextCursor"`
}

func (h *ProductsHandler) Products(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	var dto usecase.ProductsInDTO

	dto.Category = r.URL.Query().Get("category")
	dto.Seller = r.URL.Query().Get("seller")
	dto.Query = r.URL.Query().Get("query")

	cursor := r.URL.Query().Get("cursor")
	if cursor != "" {
		decodedCursor, err := decodeCursor(cursor)
		if err != nil {
			logger.Warn(err.Error())
			dto.Cursor = nil
		} else {
			dto.Cursor.ID = decodedCursor.ID
			dto.Cursor.CreatedAt = decodedCursor.CreatedAt
			dto.Cursor.Rating = decodedCursor.Rating
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

	dto.SortBy = domain.SortByType(r.URL.Query().Get("sortBy"))

	page, err := usecase.Products(
		dto, h.productRepo,
	)

	if errors.Is(err, domain.ErrInvalidLimit) {
		logger.Warn(err.Error())
	} else if errors.Is(err, domain.ErrInvalidSortBy) {
		logger.Warn(err.Error())
	} else if errors.Is(err, domain.ErrInvalidPriceRange) {
		logger.Warn(err.Error())
	} else if errors.Is(err, domain.ErrInvalidCursor) {
		logger.Warn(err.Error())
	} else if err != nil {
		logger.Error(err.Error())
		page.Products = []domain.Product{}
	}

	response := productsResponse{
		Products:   make([]productResponse, 0, len(page.Products)),
		NextCursor: "",
	}
	if page.NextCursor != nil {
		cursor, err := encodeCursor(page.NextCursor.CreatedAt,
			page.NextCursor.ID, page.NextCursor.Rating)
		if err != nil {
			logger.Error(err.Error())
		}
		response.NextCursor = cursor
	}

	// TODO: мы заранее знаем размер page.Products => можно соптимизировать

	for i, product := range page.Products {
		pictures := []string{}
		if i < len(page.PictureURLs) && page.PictureURLs[i] != nil {
			pictures = page.PictureURLs[i]
		}
		var rating float64
		if i < len(page.Ratings) {
			rating = page.Ratings[i]
		}
		var reviews uint
		if i < len(page.ReviewCounts) {
			reviews = page.ReviewCounts[i]
		}
		response.Products = append(response.Products, productResponse{
			ID: product.ID, Name: product.Name, PictureURLs: pictures,
			Price: product.Price, Discount: product.Discount,
			PriceResult: page.PricesResult[i], Rating: rating, ReviewsCount: reviews,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		logger.Error(err.Error())
	}
}
