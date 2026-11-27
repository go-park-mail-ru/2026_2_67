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

func (r *SellerRepoPostgres) SearchByQuery(query string, limit uint) ([]domain.Seller, error) {
	queryText := `WITH q AS (
    	SELECT websearch_to_tsquery('ru', $1) AS query
	)
	SELECT s.user_id, s.name, s.avatar_url
	FROM seller s, q
	WHERE s.fts @@ q.query
   		OR s.name ILIKE '%' || $1 || '%'
	ORDER BY ts_rank(s.fts, q.query) DESC, s.user_id DESC`
	args := []any{query}
	if limit != 0 {
		queryText += " LIMIT $2"
		args = append(args, limit)
	}

	rows, err := r.db.Query(queryText, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска продавцов: %w", err)
	}
	defer rows.Close()

	sellers := make([]domain.Seller, 0)
	for rows.Next() {
		var seller domain.Seller
		if err := rows.Scan(&seller.UserID, &seller.Name, &seller.AvatarURL); err != nil {
			return nil, fmt.Errorf("ошибка чтения продавца: %w", err)
		}
		sellers = append(sellers, seller)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения результатов поиска продавцов: %w", err)
	}
	return sellers, nil
}

func (r *SellerRepoPostgres) Search(params domain.SellerSearchParams) ([]domain.Seller, error) {
	conditions := make([]string, 0, 4)
	args := make([]any, 0, 6)
	addCondition := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if len(params.Category) > 0 {
		categories := make([]string, 0, len(params.Category))
		for _, category := range params.Category {
			if category != "" {
				args = append(args, category)
				categories = append(categories, fmt.Sprintf("LOWER($%d)", len(args)))
			}
		}
		if len(categories) > 0 {
			conditions = append(conditions, "LOWER(c.category) IN ("+strings.Join(categories, ", ")+")")
		}
	}
	if params.Query != "" {
		args = append(args, params.Query, "%"+params.Query+"%")
		conditions = append(conditions, fmt.Sprintf("(p.fts @@ websearch_to_tsquery('ru', $%d) OR p.name ILIKE $%d)", len(args)-1, len(args)))
	}
	if params.PriceMin != 0 {
		addCondition(`p.price::BIGINT * (100 - p.discount) / 100 >= $%d`, params.PriceMin)
	}
	if params.PriceMax != nil {
		addCondition(`p.price::BIGINT * (100 - p.discount) / 100 <= $%d`, *params.PriceMax)
	}
	if params.RatingMin != 0 {
		addCondition(`reviews.rating >= $%d`, params.RatingMin)
	}
	if params.RatingMax != nil {
		addCondition(`reviews.rating <= $%d`, *params.RatingMax)
	}
	addCondition(`p.available_count > $%d`, 0)

	args = append(args, params.Offset, params.Limit)
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := `SELECT s.user_id, s.name, s.description, s.avatar_url, s.created_at, s.updated_at
		FROM product p
		JOIN seller s ON p.seller_id = s.user_id
		JOIN category c ON p.category_id = c.id
		LEFT JOIN LATERAL (
			SELECT COALESCE(AVG(pr.rating), 0)::double precision AS rating
			FROM product_review pr WHERE pr.product_id = p.id
		) reviews ON true` +
		where +
		fmt.Sprintf(` GROUP BY s.user_id ORDER BY COUNT(s.user_id) DESC, s.name ASC, s.user_id ASC OFFSET $%d LIMIT $%d;`,
			len(args)-1, len(args))

	rows, err := r.db.Query(query, args...)
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
