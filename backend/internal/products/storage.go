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
	sampleProducts := []Product{
		{
			ID:          1,
			Name:        "Смартфон Apple iPhone 15 128GB",
			PictureURLs: []string{"https://example.com/images/iphone15_1.jpg", "https://example.com/images/iphone15_2.jpg"},
			Price:       79990,
			Rating:      4.9,
			ReviewsCount: 156,
		},
		{
			ID:          2,
			Name:        "Беспроводные наушники Sony WH-1000XM5",
			PictureURLs: []string{"https://example.com/images/sony_wh1000xm5.jpg"},
			Price:       32990,
			Rating:      4.8,
			ReviewsCount: 89,
		},
		{
			ID:          3,
			Name:        "Умная колонка Яндекс Станция Миди",
			PictureURLs: []string{"https://example.com/images/yandex_midi.jpg"},
			Price:       14990,
			Rating:      4.7,
			ReviewsCount: 230,
		},
	}

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
