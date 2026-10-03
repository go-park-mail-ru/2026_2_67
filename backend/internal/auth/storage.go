package auth

// TODO: реализовать

type Storage interface {
	SelectRefreshTokenOf(userID int64) string
	SelectUserByLoginOrEmail(loginOrEmail string) (User, bool)
	InsertRefreshToken(refreshToken RefreshToken) error
	GetRefreshTokenOf(userID int64) RefreshToken
	InsertUser(login string, email string, password string) (User, error)
}
