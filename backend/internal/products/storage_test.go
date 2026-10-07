package products

import (
	"sync"
	"testing"
)

func TestInMemoryDB_SelectAllProducts(t *testing.T) {
	db := NewInMemoryDB()

	products, err := db.SelectAllProducts()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(products) == 0 {
		t.Errorf("expected sample products, got empty list")
	}

	for _, p := range products {
		if p.ID == 0 {
			t.Errorf("expected non-zero ID for product %s", p.Name)
		}
		if p.Name == "" {
			t.Errorf("expected non-empty name for product ID %d", p.ID)
		}
	}
}

func TestInMemoryDB_InsertProduct(t *testing.T) {
	db := &InMemoryDB{
		products:      make(map[int64]Product),
		productOrder:  make([]int64, 0),
		nextProductID: 1,
	}

	newProduct := Product{
		Name:         "Тестовый товар",
		PictureURLs:  []string{"https://example.com/test.jpg"},
		Price:        1000,
		Rating:       5.0,
		ReviewsCount: 10,
	}

	inserted, err := db.InsertProduct(newProduct)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if inserted.ID != 1 {
		t.Errorf("expected ID 1, got %d", inserted.ID)
	}

	products, err := db.SelectAllProducts()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}

	if products[0].Name != newProduct.Name {
		t.Errorf("expected name %s, got %s", newProduct.Name, products[0].Name)
	}
}

func TestInMemoryDB_Concurrency(t *testing.T) {
	db := NewInMemoryDB()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				_, _ = db.InsertProduct(Product{
					Name:  "Параллельный товар",
					Price: 100 * idx,
				})
			} else {
				_, _ = db.SelectAllProducts()
			}
		}(i)
	}
	wg.Wait()
}
