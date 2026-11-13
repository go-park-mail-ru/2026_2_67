package postgres_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"vibe_market/backend/internal/pickup_points/domain"
	"vibe_market/backend/internal/pickup_points/repository/postgres"
)

func TestPickupPointRepoList(t *testing.T) {
	queryErr := errors.New("query failed")
	rowsErr := errors.New("rows failed")
	tests := []struct {
		name     string
		rows     *sqlmock.Rows
		queryErr error
		want     []domain.PickupPoint
		wantErr  error
		errText  string
	}{
		{
			name: "success",
			rows: sqlmock.NewRows([]string{"user_id", "longitude", "latitude"}).
				AddRow(int64(7), 37.6173, 55.7558).
				AddRow(int64(19), -74.006, 40.7128),
			want: []domain.PickupPoint{
				{ID: 7, Longitude: 37.6173, Latitude: 55.7558},
				{ID: 19, Longitude: -74.006, Latitude: 40.7128},
			},
		},
		{
			name: "empty",
			rows: sqlmock.NewRows([]string{"user_id", "longitude", "latitude"}),
			want: []domain.PickupPoint{},
		},
		{
			name: "query error", queryErr: queryErr, wantErr: queryErr,
		},
		{
			name:    "scan error",
			rows:    sqlmock.NewRows([]string{"user_id", "longitude", "latitude"}).AddRow(7, "invalid", 55.7558),
			errText: `sql: Scan error on column index 1, name "longitude"`,
		},
		{
			name: "rows error",
			rows: sqlmock.NewRows([]string{"user_id", "longitude", "latitude"}).
				AddRow(7, 37.6173, 55.7558).AddRow(19, 0, 0).RowError(1, rowsErr),
			wantErr: rowsErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			expectation := mock.ExpectQuery("SELECT user_id, longitude, latitude FROM pickup_point ORDER BY user_id").WithArgs()
			if tt.queryErr != nil {
				expectation.WillReturnError(tt.queryErr)
			} else {
				expectation.WillReturnRows(tt.rows).RowsWillBeClosed()
			}

			got, err := postgres.NewPickupPointRepo(db).List(context.Background())
			if tt.errText != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errText) {
					t.Fatalf("error = %v, want containing %q", err, tt.errText)
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() = %#v, want %#v", got, tt.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPickupPointRepoListCanceledContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = postgres.NewPickupPointRepo(db).List(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
