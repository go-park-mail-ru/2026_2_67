package delivery

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/auth/middleware"
	orderDomain "vibe_market/backend/internal/orders/domain"
	"vibe_market/backend/internal/products/domain"
	"vibe_market/backend/internal/products/usecase"
)

type postProductReviewRequest struct {
	Review string `json:"review"`
	Rating int    `json:"rating"`
}

type postProductReview422Response struct {
	ErrMessage string `json:"errMessage"`
}

type reviewCursorJSON struct {
	CreatedAt time.Time `json:"createdAt"`
	BuyerID   int64     `json:"buyerId"`
	Rating    int       `json:"rating"`
}

type getProductReviewItemResponse struct {
	BuyerID         int64     `json:"buyerId"`
	BuyerName       string    `json:"buyerName"`
	BuyerSurname    string    `json:"buyerSurname"`
	BuyerPatronymic string    `json:"buyerPatronymic"`
	BuyerAvatarURL  string    `json:"buyerAvatarUrl"`
	Review          string    `json:"review"`
	Rating          int       `json:"rating"`
	CreatedAt       time.Time `json:"createdAt"`
}

type getProductReviewsResponse struct {
	Reviews    []getProductReviewItemResponse `json:"reviews"`
	NextCursor string                         `json:"nextCursor"`
}

func encodeReviewCursor(cursor domain.ProductReviewCursor) string {
	b, _ := json.Marshal(reviewCursorJSON{
		CreatedAt: cursor.CreatedAt.UTC(),
		BuyerID:   cursor.BuyerID,
		Rating:    cursor.Rating,
	})
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeReviewCursor(raw string) (*domain.ProductReviewCursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, domain.ErrProductReviewCursorInvalid
	}
	var c reviewCursorJSON
	if err := json.Unmarshal(b, &c); err != nil || c.BuyerID <= 0 || c.CreatedAt.IsZero() {
		return nil, domain.ErrProductReviewCursorInvalid
	}
	return &domain.ProductReviewCursor{
		CreatedAt: c.CreatedAt,
		BuyerID:   c.BuyerID,
		Rating:    c.Rating,
	}, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (h *ProductsHandler) GetProductReviews(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || productID <= 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	query := r.URL.Query()

	var limit int
	if rawLimit := query.Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit <= 0 {
			logger.Warn(domain.ErrProductReviewLimitInvalid.Error())
			http.Error(w, domain.ErrProductReviewLimitInvalid.Error(), http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	var cursor *domain.ProductReviewCursor
	if rawCursor := query.Get("cursor"); rawCursor != "" {
		cursor, err = decodeReviewCursor(rawCursor)
		if err != nil {
			logger.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	sortByRaw := query.Get("sortBy")
	if sortByRaw == "" {
		sortByRaw = query.Get("soryBy")
	}

	out, err := h.productReviewUsecase.GetProductReviews(usecase.GetProductReviewsInDTO{
		Context:   r.Context(),
		ProductID: productID,
		SortBy:    domain.ProductReviewSortBy(sortByRaw),
		Limit:     limit,
		Cursor:    cursor,
	})
	if err != nil {
		if errors.Is(err, domain.ErrProductReviewSortByInvalid) {
			logger.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := getProductReviewsResponse{
		Reviews: make([]getProductReviewItemResponse, 0, len(out.Reviews)),
	}
	for _, item := range out.Reviews {
		response.Reviews = append(response.Reviews, getProductReviewItemResponse{
			BuyerID:         item.BuyerID,
			BuyerName:       item.BuyerName,
			BuyerSurname:    derefString(item.BuyerSurname),
			BuyerPatronymic: derefString(item.BuyerPatronymic),
			BuyerAvatarURL:  derefString(item.BuyerAvatarURL),
			Review:          derefString(item.Review),
			Rating:          item.Rating,
			CreatedAt:       item.CreatedAt.UTC(),
		})
	}
	if out.NextCursor != nil {
		response.NextCursor = encodeReviewCursor(*out.NextCursor)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(err.Error())
	}
}

func (h *ProductsHandler) PostProductReview(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())
	accessToken := r.Context().Value(middleware.AccessTokenPayloadKey).(*authDomain.AccessTokenPayload)

	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	decoder := json.NewDecoder(r.Body)

	defer r.Body.Close()

	var request postProductReviewRequest
	err = decoder.Decode(&request)
	if err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = h.productReviewUsecase.PostProductReview(usecase.PostProductReviewInDTO{
		ProductID: productID, BuyerID: accessToken.UserID, Review: &request.Review, Rating: request.Rating})

	if errors.Is(err, orderDomain.ErrOrdersNotFound) || errors.Is(err, domain.ErrProductReviewRatingInvalid) {
		encoder := json.NewEncoder(w)
		w.WriteHeader(http.StatusUnprocessableEntity)
		err = encoder.Encode(postProductReview422Response{
			ErrMessage: err.Error(),
		})
		if err != nil {
			logger.Error(err.Error())
		}
		return
	}
	if errors.Is(err, domain.ErrProductReviewAlreadyExists) {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type productReviewResponse struct {
	Rating int    `json:"rating"`
	Review string `json:"review"`
}

type patchProductReviewRequest struct {
	Rating *int   `json:"rating"`
	Review string `json:"review"`
}

func parsePositivePathID(r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func (h *ProductsHandler) GetProductReview(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())

	productID, ok := parsePositivePathID(r, "product_id")
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	buyerID, ok := parsePositivePathID(r, "buyer_id")
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	out, err := h.productReviewUsecase.GetProductReview(usecase.GetProductReviewInDTO{
		Context:   r.Context(),
		ProductID: productID,
		BuyerID:   buyerID,
	})
	if errors.Is(err, domain.ErrProductReviewNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(productReviewResponse{
		Rating: out.Rating,
		Review: derefString(out.Review),
	}); err != nil {
		logger.Error(err.Error())
	}
}

func (h *ProductsHandler) PatchProductReview(w http.ResponseWriter, r *http.Request) {
	logger := h.logger(r.Context())
	accessToken, ok := r.Context().Value(middleware.AccessTokenPayloadKey).(*authDomain.AccessTokenPayload)
	if !ok || accessToken == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	productID, ok := parsePositivePathID(r, "product_id")
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	buyerID, ok := parsePositivePathID(r, "buyer_id")
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if accessToken.UserID != buyerID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	defer r.Body.Close()
	var request patchProductReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if request.Rating == nil {
		logger.Warn("не указан rating")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	out, err := h.productReviewUsecase.PatchProductReview(usecase.PatchProductReviewInDTO{
		Context:   r.Context(),
		ProductID: productID,
		BuyerID:   buyerID,
		Review:    &request.Review,
		Rating:    *request.Rating,
	})
	if errors.Is(err, domain.ErrProductReviewNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if errors.Is(err, orderDomain.ErrOrdersNotFound) || errors.Is(err, domain.ErrProductReviewRatingInvalid) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		if encErr := json.NewEncoder(w).Encode(postProductReview422Response{
			ErrMessage: err.Error(),
		}); encErr != nil {
			logger.Error(encErr.Error())
		}
		return
	}
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(productReviewResponse{
		Rating: out.Rating,
		Review: derefString(out.Review),
	}); err != nil {
		logger.Error(err.Error())
	}
}
