package repo

import (
	"sync"
	"time"
	"vibe_market/backend/internal/auth/domain"
)

// ==========================================
// 1. UserInMemoryDB (Реализация UserRepo)
// ==========================================

type UserInMemoryDB struct {
	mu           sync.RWMutex
	users        map[int64]domain.User
	usersByEmail map[string]int64 // Индекс для проверки уникальности и поиска по email
	nextUserID   int64
}

func NewUserInMemoryDB() *UserInMemoryDB {
	return &UserInMemoryDB{
		users:        make(map[int64]domain.User),
		usersByEmail: make(map[string]int64),
		nextUserID:   1,
	}
}

func (db *UserInMemoryDB) Insert(email string, password string) (domain.User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.usersByEmail[email]; exists {
		return domain.User{}, domain.ErrEmailAlreadyExists
	}

	now := time.Now()
	user := domain.User{
		UserID:       db.nextUserID,
		Email:        email,
		PasswordHash: password,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	db.users[db.nextUserID] = user
	db.usersByEmail[email] = db.nextUserID
	db.nextUserID++

	return user, nil
}

func (db *UserInMemoryDB) DeleteByUserID(userID int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	user, exists := db.users[userID]
	if !exists {
		return nil
	}

	delete(db.usersByEmail, user.Email)
	delete(db.users, userID)
	return nil
}

func (db *UserInMemoryDB) SelectByEmail(email string) (domain.User, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	userID, exists := db.usersByEmail[email]
	if !exists {
		return domain.User{}, false
	}

	user, exists := db.users[userID]
	return user, exists
}

func (db *UserInMemoryDB) SelectByUserID(userID int64) (domain.User, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, exists := db.users[userID]
	return user, exists
}

func (db *UserInMemoryDB) TruncateByUserID(userID int64) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	user, exists := db.users[userID]
	if !exists {
		return false
	}

	delete(db.usersByEmail, user.Email)
	delete(db.users, userID)
	return true
}

// ==========================================
// 2. RefreshTokenInMemoryDB (Реализация RefreshTokenRepo)
// ==========================================

type RefreshTokenInMemoryDB struct {
	mu            sync.RWMutex
	refreshTokens map[string]domain.RefreshToken // Ключ — TokenHash
}

func NewRefreshTokenInMemoryDB() *RefreshTokenInMemoryDB {
	return &RefreshTokenInMemoryDB{
		refreshTokens: make(map[string]domain.RefreshToken),
	}
}

func (db *RefreshTokenInMemoryDB) Insert(t domain.RefreshToken) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.refreshTokens[t.TokenHash] = t
	return nil
}

func (db *RefreshTokenInMemoryDB) Revoke(userID int64, deviceName string) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	deleted := false
	for hash, token := range db.refreshTokens {
		if token.UserID == userID && token.DeviceName == deviceName {
			delete(db.refreshTokens, hash)
			deleted = true
		}
	}
	return deleted
}

func (db *RefreshTokenInMemoryDB) SelectByHashAndDeviceName(hash string, deviceName string) (domain.RefreshToken, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	token, exists := db.refreshTokens[hash]
	if !exists || token.DeviceName != deviceName {
		return domain.RefreshToken{}, false
	}

	return token, true
}
