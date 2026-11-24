package postgres

import (
	"database/sql"
	"fmt"
	"vibe_market/backend/internal/categories/domain"
)

type CategoryRepoPostgres struct {
	db *sql.DB
}

func NewCategoryRepo(db *sql.DB) *CategoryRepoPostgres {
	return &CategoryRepoPostgres{db: db}
}

func (repo *CategoryRepoPostgres) Search(params domain.CategoriesSearchParams) ([]domain.Category, error) {
	if params.SortBy != domain.SortByNew && params.SortBy != domain.SortByPopular {
		return []domain.Category{}, domain.ErrSortByInvalid
	}

	result := make([]domain.Category, 0)
	var query string

	switch params.SortBy {
	case domain.SortByNew:
		query = "SELECT id, category, picture_url FROM category ORDER BY id DESC"

	case domain.SortByPopular:
		query = `
		SELECT c.id, c.category, c.picture_url FROM category c LEFT JOIN product p ON p.category_id = c.id
		GROUP BY c.id, c.category, c.picture_url ORDER BY COUNT(p.id) DESC, c.id DESC
		`
	}

	var (
		rows *sql.Rows
		err  error
	)
	if params.Limit != 0 {
		query += " LIMIT $1;"
		rows, err = repo.db.Query(query, params.Limit)
	} else {
		query += ";"
		rows, err = repo.db.Query(query)
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
		WHERE fts @@ websearch_to_tsquery('ru', $1)
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
