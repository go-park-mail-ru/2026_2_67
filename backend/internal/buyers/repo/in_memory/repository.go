package repo

import (
	"sync"
	userDomain "vibe_market/backend/internal/auth/domain"
	buyerDomain "vibe_market/backend/internal/buyers/domain"
)

type BuyerInMemoryDB struct {
	sync.RWMutex
	buyers map[int64]buyerDomain.Buyer
}

func NewBuyerInMemoryDB(userRepo userDomain.UserRepo) *BuyerInMemoryDB {
	return &BuyerInMemoryDB{
		buyers: make(map[int64]buyerDomain.Buyer),
	}
}

func (repo *BuyerInMemoryDB) SelectByUserID(userID int64) (buyerDomain.Buyer, bool) {
	repo.RLock()
	defer repo.RUnlock()

	buyer, ok := repo.buyers[userID]
	return buyer, ok
}

func (repo *BuyerInMemoryDB) Insert(buyer buyerDomain.Buyer) bool {
	repo.Lock()
	defer repo.Unlock()

	_, ok := repo.buyers[buyer.UserID]
	if ok {
		return false
	}
	repo.buyers[buyer.UserID] = buyer
	return true
}
