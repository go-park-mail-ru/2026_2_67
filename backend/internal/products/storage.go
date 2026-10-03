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
			PictureURLs: []string{"https://avatars.mds.yandex.net/get-mpic/15259477/2a000001957a3df8caedd9c7aa21b8a42e19/orig", "https://avatars.mds.yandex.net/get-mpic/21171048/picee500b1138189ee3037121a43929d45c/orig"},
			Price:       79990,
			Rating:      4.9,
			ReviewsCount: 156,
		},
		{
			ID:          2,
			Name:        "Беспроводные наушники Sony WH-1000XM5",
			PictureURLs: []string{"https://avatars.mds.yandex.net/get-mpic/16418886/2a00000196d8d1202190e2a2a507cc5a0267/orig"},
			Price:       32990,
			Rating:      4.8,
			ReviewsCount: 89,
		},
		{
			ID:          3,
			Name:        "Умная колонка Яндекс Станция Миди",
			PictureURLs: []string{"https://avatars.mds.yandex.net/i?id=d28b20934449864f1b2e39ca5b5ded5b_l-5226953-images-thumbs&n=13"},
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
