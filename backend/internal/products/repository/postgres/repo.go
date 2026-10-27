package postgres

import "database/sql"

type ProductRepoPostgres struct {
	db *sql.DB
}

func NewProductRepo(db *sql.DB) *ProductRepoPostgres {
	return &ProductRepoPostgres{db: db}
}
