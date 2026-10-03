package auth

// TODO: реализовать

import (
	"errors"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	SelectUserByLoginOrEmail(loginOrEmail string) (User, bool)
	InsertRefreshToken(refreshToken RefreshToken) error
	InsertUser(login string, email string, password string) (User, error)
}

var (
	ErrUserAlreadyExists = errors.New("пользователь с указанным login или email уже существует")
)

type InMemoryDB struct {
	sync.RWMutex
	users           map[int64]User
	usersByIdentity map[string]int64 // key: login или email (в нижнем регистре), value: userID

	refreshTokens []RefreshToken

	nextUserID int64
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		users:           make(map[int64]User),
		usersByIdentity: make(map[string]int64),
		refreshTokens:   make([]RefreshToken, 0),
		nextUserID:      1,
	}
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

// InsertUser создает нового пользователя с ролью по умолчанию (RoleBuyer).
func (db *InMemoryDB) InsertUser(login string, email string, password string) (User, error) {
	db.Lock()
	defer db.Unlock()

	cleanLogin := strings.ToLower(login)
	cleanEmail := strings.ToLower(email)

	// Проверяем уникальность логина и email
	if _, exists := db.usersByIdentity[cleanLogin]; exists {
		return User{}, ErrUserAlreadyExists
	}
	if _, exists := db.usersByIdentity[cleanEmail]; exists {
		return User{}, ErrUserAlreadyExists
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

	// Сохраняем пользователя и обновляем индексы
	db.users[user.UserID] = user
	db.usersByIdentity[cleanLogin] = user.UserID
	db.usersByIdentity[cleanEmail] = user.UserID

	db.nextUserID++

	return user, nil
}

// InsertRefreshToken сохраняет refresh token в память.
func (db *InMemoryDB) InsertRefreshToken(refreshToken RefreshToken) error {
	db.Lock()
	defer db.Unlock()

	db.refreshTokens = append(db.refreshTokens, refreshToken)
	return nil
}
