package postgres

import (
	"context"
	"database/sql"

	"vibe_market/backend/internal/pickup_points/domain"
)

type PickupPointRepoPostgres struct {
	db *sql.DB
}

func NewPickupPointRepo(db *sql.DB) *PickupPointRepoPostgres {
	return &PickupPointRepoPostgres{db: db}
}

func (repo *PickupPointRepoPostgres) List(ctx context.Context) ([]domain.PickupPoint, error) {
	rows, err := repo.db.QueryContext(ctx, "SELECT user_id, longitude, latitude FROM pickup_point ORDER BY user_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]domain.PickupPoint, 0)
	for rows.Next() {
		var point domain.PickupPoint
		if err := rows.Scan(&point.ID, &point.Longitude, &point.Latitude); err != nil {
			return nil, err
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return points, nil
}
