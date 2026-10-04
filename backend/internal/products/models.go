package products

type Product struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	PictureURLs  []string `json:"pictureUrls"`
	Price        int      `json:"price"`
	Rating       float64  `json:"rating"`
	ReviewsCount int      `json:"reviewsCount"`
}
