package postgres

import (
	"database/sql"
	"vibe_market/backend/internal/products/domain"
)

func (repo *ProductRepoPostgres) SearchCategories(params domain.CategoriesSearchParams) ([]domain.ProductCategory, error) {
	if params.Limit < 0 {
		return []domain.ProductCategory{}, domain.ErrInvalidLimit
	}
	if params.SortBy == "" {
		params.SortBy = domain.SortByNew
	}
	if params.SortBy != domain.SortByNew && params.SortBy != domain.SortByPopular {
		return []domain.ProductCategory{}, domain.ErrInvalidSortBy
	}

	result := make([]domain.ProductCategory, 0)
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
		return []domain.ProductCategory{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var category domain.ProductCategory
		err = rows.Scan(&category.ID, &category.Category, &category.PictureURL)
		if err != nil {
			return []domain.ProductCategory{}, err
		}
		result = append(result, category)
	}

	if rows.Err() != nil {
		return []domain.ProductCategory{}, rows.Err()
	}

	return result, nil
}
