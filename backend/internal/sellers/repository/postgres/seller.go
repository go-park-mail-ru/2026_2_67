package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"vibe_market/backend/internal/sellers/domain"
)

type SellerRepoPostgres struct {
	db *sql.DB
}

func NewSellerRepo(db *sql.DB) *SellerRepoPostgres {
	return &SellerRepoPostgres{db: db}
}

const DefaultSellerLimit = 10
const MaxSellerLimit = 30

func (repo *SellerRepoPostgres) Search(params domain.SellerSearchParams) ([]domain.Seller, error) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 4)
	addCondition := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if params.Category != "" {
		addCondition(`LOWER(c.category) = LOWER($%d)`, params.Category)
	}
	if params.Query != "" {
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(params.Query)
		addCondition(`p.name ILIKE $%d ESCAPE E'\\'`, escaped+"%")
	}
	if params.Limit == 0 {
		params.Limit = DefaultSellerLimit
	}
	params.Limit = min(MaxSellerLimit, params.Limit)

	args = append(args, params.Offset, params.Limit)
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := `SELECT s.user_id, s.name, s.description, s.avatar_url, s.created_at, s.updated_at
		FROM product p
		JOIN seller s ON p.seller_id = s.user_id
		JOIN product_category c ON p.category_id = c.id` +
		where +
		fmt.Sprintf(` GROUP BY s.user_id ORDER BY COUNT(s.user_id) DESC, s.name ASC OFFSET $%d LIMIT $%d;`,
			len(args)-1, len(args))

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return []domain.Seller{}, err
	}

	defer rows.Close()

	sellers := make([]domain.Seller, 0)
	for rows.Next() {
		var seller domain.Seller

		err = rows.Scan(&seller.UserID, &seller.Name, &seller.Description, &seller.AvatarURL,
			&seller.CreatedAt, &seller.UpdatedAt)
		if err != nil {
			return []domain.Seller{}, err
		}
		sellers = append(sellers, seller)
	}

	if rows.Err() != nil {
		return []domain.Seller{}, rows.Err()
	}

	return sellers, nil
}
