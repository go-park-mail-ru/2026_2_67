package postgres

import (
	"database/sql"
	"vibe_market/backend/internal/categories/domain"
)

type CategoryRepoPostgres struct {
	db *sql.DB
}

func NewCategoryRepo(db *sql.DB) *CategoryRepoPostgres {
	return &CategoryRepoPostgres{db: db}
}

func (repo *CategoryRepoPostgres) Search(params domain.CategoriesSearchParams) ([]domain.Category, error) {
	if params.SortBy == "" {
		params.SortBy = domain.SortByNew
	}
	if params.SortBy != domain.SortByNew && params.SortBy != domain.SortByPopular {
		return []domain.Category{}, domain.ErrInvalidSortBy
	}

	result := make([]domain.Category, 0)
	var query string

	switch params.SortBy {
	case domain.SortByNew:
		query = "SELECT id, category, picture_url FROM product_category ORDER BY id DESC"

	case domain.SortByPopular:
		query = `
		SELECT c.id, c.category, c.picture_url FROM product_category c LEFT JOIN product p ON p.category_id = c.id
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
