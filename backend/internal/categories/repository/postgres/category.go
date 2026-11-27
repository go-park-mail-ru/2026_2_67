package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"vibe_market/backend/internal/categories/domain"
)

type CategoryRepoPostgres struct {
	db *sql.DB
}

func NewCategoryRepo(db *sql.DB) *CategoryRepoPostgres {
	return &CategoryRepoPostgres{db: db}
}

func (repo *CategoryRepoPostgres) Search(params domain.CategoriesSearchParams) ([]domain.Category, error) {
	switch params.SortBy {
	case domain.SortByNew, domain.SortByPopular:
	default:
		return []domain.Category{}, domain.ErrSortByInvalid
	}

	result := make([]domain.Category, 0)
	var query string
	args := make([]any, 0, 8)
	conditions := make([]string, 0, 6)
	addCondition := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if params.Query != "" {
		args = append(args, params.Query, "%"+params.Query+"%")
		conditions = append(conditions, fmt.Sprintf("(p.fts @@ websearch_to_tsquery('ru', $%d) OR p.name ILIKE $%d)", len(args)-1, len(args)))
	}
	if len(params.Seller) > 0 {
		placeholders := make([]string, 0, len(params.Seller))
		for _, seller := range params.Seller {
			if seller == "" {
				continue
			}
			args = append(args, strings.ToLower(seller))
			placeholders = append(placeholders, fmt.Sprintf("LOWER(s.name) = $%d", len(args)))
		}
		if len(placeholders) > 0 {
			conditions = append(conditions, "("+strings.Join(placeholders, " OR ")+")")
		}
	}
	if params.PriceMin != 0 {
		addCondition("p.price::BIGINT * (100 - p.discount) / 100 >= $%d", params.PriceMin)
	}
	if params.PriceMax != nil {
		addCondition("p.price::BIGINT * (100 - p.discount) / 100 <= $%d", *params.PriceMax)
	}
	if params.RatingMin > 0 {
		addCondition("reviews.rating >= $%d", params.RatingMin)
	}
	if params.RatingMax != nil {
		addCondition("reviews.rating <= $%d", *params.RatingMax)
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	if len(conditions) == 0 {
		if params.SortBy == domain.SortByPopular {
			query = `SELECT c.id, c.category, c.picture_url FROM category c LEFT JOIN product p ON p.category_id = c.id
			GROUP BY c.id, c.category, c.picture_url ORDER BY COUNT(p.id) DESC, c.id DESC `
		} else {
			query = "SELECT id, category, picture_url FROM category ORDER BY id DESC"
		}
	} else {
		query = `SELECT c.id, c.category, c.picture_url
			FROM product p
			JOIN category c ON p.category_id = c.id
			JOIN seller s ON p.seller_id = s.user_id
			LEFT JOIN LATERAL (
				SELECT COALESCE(AVG(pr.rating), 0)::double precision AS rating
				FROM product_review pr WHERE pr.product_id = p.id
			) reviews ON true` + where +
			` GROUP BY c.id, c.category, c.picture_url`
		if params.SortBy == domain.SortByPopular {
			query += " ORDER BY COUNT(p.id) DESC, c.id DESC"
		} else {
			query += " ORDER BY c.id DESC"
		}
	}

	var (
		rows *sql.Rows
		err  error
	)
	if params.Limit != 0 {
		args = append(args, params.Limit)
		query += fmt.Sprintf(" LIMIT $%d;", len(args))
		rows, err = repo.db.Query(query, args...)
	} else {
		query += ";"
		rows, err = repo.db.Query(query, args...)
	}
	if err != nil {
		return []domain.Category{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var category domain.Category
		err = rows.Scan(&category.ID, &category.Category, &category.PictureURL)
		if err != nil {
			return []domain.Category{}, err
		}
		result = append(result, category)
	}

	if rows.Err() != nil {
		return []domain.Category{}, rows.Err()
	}

	return result, nil
}

func (repo *CategoryRepoPostgres) SearchByQuery(query string, limit uint) ([]domain.Category, error) {
	queryText := `
		SELECT id, category, picture_url
		FROM category
		WHERE fts @@ websearch_to_tsquery('ru', $1) OR category ILIKE '%' || $1 || '%'
		ORDER BY ts_rank(fts, websearch_to_tsquery('ru', $1)) DESC, id DESC`
	args := []any{query}
	if limit != 0 {
		queryText += " LIMIT $2"
		args = append(args, limit)
	}

	rows, err := repo.db.Query(queryText, args...)
	if err != nil {
		return []domain.Category{}, fmt.Errorf("search categories by query: %w", err)
	}
	defer rows.Close()

	categories := make([]domain.Category, 0)
	for rows.Next() {
		var category domain.Category
		if err := rows.Scan(&category.ID, &category.Category, &category.PictureURL); err != nil {
			return []domain.Category{}, fmt.Errorf("scan category search result: %w", err)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return []domain.Category{}, fmt.Errorf("read category search results: %w", err)
	}
	return categories, nil
}
