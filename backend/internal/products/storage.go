package products

import (
	"sync"
)

type Storage interface {
	SelectAllProducts() ([]Product, error)
	InsertProduct(product Product) (Product, error)
}

type InMemoryDB struct {
	sync.RWMutex
	products      map[int64]Product
	productOrder  []int64
	nextProductID int64
}

func NewInMemoryDB() *InMemoryDB {
	db := &InMemoryDB{
		products:      make(map[int64]Product),
		productOrder:  make([]int64, 0),
		nextProductID: 1,
	}

	db.initSampleProducts()

	return db
}

func (db *InMemoryDB) initSampleProducts() {
	// Initialize sample products using paths relative to the frontend's public directory
	sampleProducts := []Product{
		{
			ID:           1,
			Name:         "Смартфон Apple iPhone 15 128GB",
			PictureURLs:  []string{"/images/smartphone-1.jpg", "/images/smartphone-2.jpg"},
			Price:        79990,
			Rating:       4.9,
			ReviewsCount: 156,
		},
		{
			ID:           2,
			Name:         "Беспроводные наушники Sony WH-1000XM5",
			PictureURLs:  []string{"/images/headphones-1.jpg", "/images/headphones-2.jpg"},
			Price:        32990,
			Rating:       4.8,
			ReviewsCount: 89,
		},
		{
			ID:           3,
			Name:         "Умная колонка Яндекс Станция Миди",
			PictureURLs:  []string{"/images/speaker-1.jpg"},
			Price:        14990,
			Rating:       4.7,
			ReviewsCount: 230,
		},
		{
			ID:           4,
			Name:         "Городской рюкзак Xiaomi Mi City Backpack",
			PictureURLs:  []string{"/images/backpack-1.jpg"},
			Price:        3490,
			Rating:       4.6,
			ReviewsCount: 42,
		},
		{
			ID:           5,
			Name:         "Беззеркальная камера Sony Alpha 6400",
			PictureURLs:  []string{"/images/camera-1.jpg"},
			Price:        85990,
			Rating:       4.9,
			ReviewsCount: 112,
		},
		{
			ID:           6,
			Name:         "Механическая клавиатура Keychron K2",
			PictureURLs:  []string{"/images/keyboard-1.jpg"},
			Price:        8990,
			Rating:       4.8,
			ReviewsCount: 205,
		},
		{
			ID:           7,
			Name:         "Умная настольная лампа Xiaomi Mi Smart LED Desk Lamp Pro",
			PictureURLs:  []string{"/images/lamp-1.jpg", "/images/lamp-2.jpg"},
			Price:        4990,
			Rating:       4.7,
			ReviewsCount: 76,
		},
		{
			ID:           8,
			Name:         "Ноутбук Apple MacBook Air 13\" M2",
			PictureURLs:  []string{"/images/laptop-1.jpg", "/images/laptop-2.jpg"},
			Price:        114990,
			Rating:       4.9,
			ReviewsCount: 340,
		},
		{
			ID:           9,
			Name:         "Умные часы Apple Watch Series 9",
			PictureURLs:  []string{"/images/smartwatch-1.jpg"},
			Price:        45990,
			Rating:       4.8,
			ReviewsCount: 184,
		},
	}

	// Populate the in-memory database with the updated products
	for _, p := range sampleProducts {
		db.products[p.ID] = p
		db.productOrder = append(db.productOrder, p.ID)

		if p.ID >= db.nextProductID {
			db.nextProductID = p.ID + 1
		}
	}
}

// SelectAllProducts возвращает список всех товаров.
func (db *InMemoryDB) SelectAllProducts() ([]Product, error) {
	db.RLock()
	defer db.RUnlock()

	result := make([]Product, 0, len(db.productOrder))
	for _, id := range db.productOrder {
		if p, ok := db.products[id]; ok {
			result = append(result, p)
		}
	}

	return result, nil
}

// InsertProduct добавляет или обновляет товар в InMemoryDB.
func (db *InMemoryDB) InsertProduct(product Product) (Product, error) {
	db.Lock()
	defer db.Unlock()

	if product.ID == 0 {
		product.ID = db.nextProductID
		db.nextProductID++
	} else if product.ID >= db.nextProductID {
		db.nextProductID = product.ID + 1
	}

	if _, exists := db.products[product.ID]; !exists {
		db.productOrder = append(db.productOrder, product.ID)
	}
	db.products[product.ID] = product

	return product, nil
}
