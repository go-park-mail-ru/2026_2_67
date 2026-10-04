package auth

// TODO: реализовать

import (
	"errors"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	SelectUserByID(userID int64) (User, bool)
	SelectUserByLoginOrEmail(loginOrEmail string) (User, bool)
	SelectRefreshTokenByHash(refreshTokenHash string) (RefreshToken, bool)
	InsertRefreshToken(refreshToken RefreshToken) error
	InsertUser(login string, email string, password string) (User, error)
	DropRefreshToken(userID int64) bool
}

var (
	errLoginAlreadyExists = errors.New("пользователь с указанным логином уже существует")
	errEmailAlreadyExists = errors.New("пользователь с указанным email уже существует")
)

type InMemoryDB struct {
	sync.RWMutex
	users           map[int64]User
	usersByIdentity map[string]int64 // key: login или email (в нижнем регистре), value: userID

	refreshTokens             map[string]RefreshToken
	refreshTokenHashsByUserID map[int64]string

	nextUserID int64
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		users:                     make(map[int64]User),
		usersByIdentity:           make(map[string]int64),
		refreshTokens:             make(map[string]RefreshToken),
		refreshTokenHashsByUserID: make(map[int64]string),
		nextUserID:                1,
	}
}

// SelectUserByID ищет пользователя по ID
func (db *InMemoryDB) SelectUserByID(userID int64) (User, bool) {
	db.RLock()
	defer db.RUnlock()

	user, ok := db.users[userID]
	return user, ok
}

// SelectUserByLoginOrEmail ищет пользователя по логину или email.
func (db *InMemoryDB) SelectUserByLoginOrEmail(loginOrEmail string) (User, bool) {
	db.RLock()
	defer db.RUnlock()

	key := strings.ToLower(loginOrEmail)
	userID, exists := db.usersByIdentity[key]
	if !exists {
		return User{}, false
	}

	user, ok := db.users[userID]
	return user, ok
}

func (db *InMemoryDB) SelectRefreshTokenByHash(refreshTokenHash string) (RefreshToken, bool) {
	db.RLock()
	defer db.RUnlock()

	refreshToken, ok := db.refreshTokens[refreshTokenHash]
	return refreshToken, ok
}

// InsertUser создает нового пользователя с ролью по умолчанию (RoleBuyer).
func (db *InMemoryDB) InsertUser(login string, email string, password string) (User, error) {
	db.Lock()
	defer db.Unlock()

	cleanLogin := strings.ToLower(login)
	cleanEmail := strings.ToLower(email)

	// Проверяем уникальность логина и email
	if _, exists := db.usersByIdentity[cleanLogin]; exists {
		return User{}, errLoginAlreadyExists
	}
	if _, exists := db.usersByIdentity[cleanEmail]; exists {
		return User{}, errEmailAlreadyExists
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	passwordHash := string(hashedBytes)

	user := User{
		UserID:       db.nextUserID,
		Login:        login,
		Role:         RoleBuyer,
		passwordHash: passwordHash,
	}

	db.users[user.UserID] = user
	db.usersByIdentity[cleanLogin] = user.UserID
	db.usersByIdentity[cleanEmail] = user.UserID

	db.nextUserID++

	return user, nil
}

// InsertRefreshToken сохраняет refresToken в память
func (db *InMemoryDB) InsertRefreshToken(refreshToken RefreshToken) error {
	db.Lock()
	defer db.Unlock()

	db.refreshTokens[refreshToken.TokenHash] = refreshToken
	db.refreshTokenHashsByUserID[refreshToken.UserID] = refreshToken.TokenHash
	return nil
}

// DropRefreshToken удаляет refreshToken из бд
func (db *InMemoryDB) DropRefreshToken(userID int64) bool {
	db.Lock()
	defer db.Unlock()

	refreshTokenHash, ok := db.refreshTokenHashsByUserID[userID]

	if !ok {
		return false
	}

	delete(db.refreshTokens, refreshTokenHash)
	delete(db.refreshTokenHashsByUserID, userID)

	return true
}
