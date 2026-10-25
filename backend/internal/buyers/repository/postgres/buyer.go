package postgres

import (
	"database/sql"
	"errors"
	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/buyers/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

type BuyerRepoPostgres struct {
	db *sql.DB
}

func NewBuyerRepo(db *sql.DB) *BuyerRepoPostgres {
	return &BuyerRepoPostgres{db}
}

func (repo *BuyerRepoPostgres) SelectByUserID(userID int64) (domain.Buyer, error) {
	row := repo.db.QueryRow(
		"SELECT user_id, name, surname, birth_date, description, avatar_url, telephone, created_at, updated_at FROM buyer WHERE user_id = $1;",
		userID)
	buyer := domain.Buyer{}

	if row.Err() != nil {
		return buyer, row.Err()
	}

	err := row.Scan(&buyer.UserID, &buyer.Name, &buyer.Surname, &buyer.BirthDate, &buyer.Description, &buyer.AvatarURL,
		&buyer.Telephone, &buyer.CreatedAt, &buyer.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return buyer, authDomain.ErrUserIDNotFound
		}
		return buyer, err
	}

	return buyer, nil
}

func (repo *BuyerRepoPostgres) Insert(buyer domain.Buyer) error {
	row := repo.db.QueryRow(
		"INSERT INTO buyer (user_id, name, surname, birth_date, description, avatar_url, telephone) VALUES ($1, $2, $3, $4, $5, $6, $7);",
		buyer.UserID, buyer.Name, buyer.Surname, buyer.BirthDate, buyer.Description, buyer.AvatarURL, buyer.Telephone,
	)
	if row.Err() != nil {
		var pgErr *pgconn.PgError
		if errors.As(row.Err(), &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUserIDAlreadyExists
		}
		return row.Err()
	}
	return nil
}

func (repo *BuyerRepoPostgres) SelectByUserIDWithUser(userID int64) (domain.Buyer, authDomain.User, error) {
	row := repo.db.QueryRow(
		`SELECT "user".id, "user".email, "user".password_hash, "user".created_at, "user".updated_at, `+
			`buyer.name, buyer.surname, buyer.birth_date, buyer.description, buyer.avatar_url, buyer.telephone, buyer.created_at, buyer.updated_at `+
			`FROM buyer LEFT JOIN "user" ON buyer.user_id = "user".id WHERE "user".id = $1;`, userID)

	buyer := domain.Buyer{}
	user := authDomain.User{}

	if row.Err() != nil {
		return buyer, user, row.Err()
	}

	err := row.Scan(&user.UserID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
		&buyer.Name, &buyer.Surname, &buyer.BirthDate, &buyer.Description, &buyer.AvatarURL, &buyer.Telephone, &buyer.CreatedAt, &buyer.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return buyer, user, authDomain.ErrUserIDNotFound
		}
		return buyer, user, err
	}
	buyer.UserID = user.UserID
	return buyer, user, nil
}
