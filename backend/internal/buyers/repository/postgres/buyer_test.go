package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/buyers/domain"
)

const (
	selectBuyerSQL         = "SELECT user_id, name, surname, birth_date, description, avatar_url, telephone, created_at, updated_at FROM buyer WHERE user_id = $1;"
	selectBuyerWithUserSQL = `SELECT "user".id, "user".email, "user".password_hash, "user".created_at, "user".updated_at,
		buyer.name, buyer.surname, buyer.birth_date, buyer.description, buyer.avatar_url, buyer.telephone, buyer.created_at, buyer.updated_at
		FROM buyer LEFT JOIN "user" ON buyer.user_id = "user".id WHERE "user".id = $1;`
	insertBuyerSQL = "INSERT INTO buyer (user_id, name, surname, birth_date, description, avatar_url, telephone) VALUES ($1, $2, $3, $4, $5, $6, $7);"
	updateBuyerSQL = `UPDATE buyer SET name = $1, surname = $2, birth_date = $3, description = $4,
		avatar_url = $5, telephone = $6 WHERE user_id = $7`
)

func newBuyerRepoTest(t *testing.T) (*BuyerRepoPostgres, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		_ = db.Close()
	})
	return NewBuyerRepo(db), mock
}

func sampleBuyer(full bool) domain.Buyer {
	buyer := domain.Buyer{
		UserID: 42, Name: "Buyer",
		CreatedAt: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC),
	}
	if full {
		surname, description, avatar, telephone := "Surname", "Description", "/avatar.png", "+79991234567"
		birthDate := time.Date(1990, 2, 3, 0, 0, 0, 0, time.UTC)
		buyer.Surname, buyer.BirthDate, buyer.Description = &surname, &birthDate, &description
		buyer.AvatarURL, buyer.Telephone = &avatar, &telephone
	}
	return buyer
}

func buyerValues(b domain.Buyer) []driver.Value {
	return []driver.Value{b.Name, b.Surname, b.BirthDate, b.Description, b.AvatarURL, b.Telephone, b.CreatedAt, b.UpdatedAt}
}

func buyerRows(joined bool) *sqlmock.Rows {
	columns := []string{"name", "surname", "birth_date", "description", "avatar_url", "telephone", "created_at", "updated_at"}
	if joined {
		columns = append([]string{"id", "email", "password_hash", "created_at", "updated_at"}, columns...)
	} else {
		columns = append([]string{"user_id"}, columns...)
	}
	return sqlmock.NewRows(columns)
}

func beginBuyerTestTx(t *testing.T, repo *BuyerRepoPostgres, mock sqlmock.Sqlmock) context.Context {
	t.Helper()
	mock.ExpectBegin()
	ctx, err := repo.BeginTx()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		mock.ExpectRollback()
		repo.RollbackTx(ctx)
	})
	return ctx
}

func TestNewBuyerRepo(t *testing.T) {
	for _, db := range []*sql.DB{new(sql.DB), nil} {
		repo := NewBuyerRepo(db)
		if repo == nil || repo.db != db {
			t.Fatalf("NewBuyerRepo(%p) = %#v, want repository retaining supplied DB", db, repo)
		}
	}
}

func TestBuyerSelect(t *testing.T) {
	for _, joined := range []bool{false, true} {
		for _, name := range []string{"full", "nullable", "no rows", "query error", "scan error", "row error"} {
			t.Run(fmt.Sprintf("joined=%t/%s", joined, name), func(t *testing.T) {
				repo, mock := newBuyerRepoTest(t)
				buyer := sampleBuyer(name != "nullable")
				user := authDomain.User{
					UserID: buyer.UserID, Email: "buyer@example.com", PasswordHash: "hash",
					CreatedAt: buyer.CreatedAt.Add(-time.Hour), UpdatedAt: buyer.UpdatedAt.Add(-time.Hour),
				}
				values := append([]driver.Value{buyer.UserID}, buyerValues(buyer)...)
				query := selectBuyerSQL
				if joined {
					query = selectBuyerWithUserSQL
					values = append([]driver.Value{user.UserID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt}, buyerValues(buyer)...)
				}
				expect := mock.ExpectQuery(query).WithArgs(buyer.UserID)
				rows := buyerRows(joined)
				var wantErr error
				switch name {
				case "query error":
					wantErr = errors.New("query failed")
					expect.WillReturnError(wantErr)
				case "no rows":
					wantErr = authDomain.ErrUserIDNotFound
					expect.WillReturnRows(rows).RowsWillBeClosed()
				default:
					if name == "scan error" {
						values[0] = "invalid ID"
					}
					rows.AddRow(values...)
					if name == "row error" {
						wantErr = errors.New("row fetch failed")
						rows.RowError(0, wantErr)
					}
					expect.WillReturnRows(rows).RowsWillBeClosed()
				}
				var got domain.Buyer
				var gotUser authDomain.User
				var err error
				if joined {
					got, gotUser, err = repo.SelectByUserIDWithUser(buyer.UserID)
				} else {
					got, err = repo.SelectByUserID(buyer.UserID)
				}
				if name == "scan error" {
					if err == nil || !strings.Contains(err.Error(), "column index 0") || !strings.Contains(err.Error(), "invalid ID") {
						t.Fatalf("error = %v, want ID scan conversion error", err)
					}
				} else if !errors.Is(err, wantErr) {
					t.Fatalf("error = %v, want %v", err, wantErr)
				}
				if name != "full" && name != "nullable" {
					buyer, user = domain.Buyer{}, authDomain.User{}
				}
				if !reflect.DeepEqual(got, buyer) {
					t.Fatalf("buyer = %#v, want %#v", got, buyer)
				}
				if joined && gotUser != user {
					t.Fatalf("user = %#v, want %#v", gotUser, user)
				}
			})
		}
	}
}

func TestBuyerBeginTx(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newBuyerRepoTest(t)
		ctx := beginBuyerTestTx(t, repo, mock)
		if tx, ok := ctx.Value(authDomain.ContextTxKey).(*sql.Tx); !ok || tx == nil {
			t.Fatalf("transaction = %v, want *sql.Tx", ctx.Value(authDomain.ContextTxKey))
		}
		if ctx.Err() != nil {
			t.Fatalf("context error = %v", ctx.Err())
		}
	})
	t.Run("error", func(t *testing.T) {
		repo, mock := newBuyerRepoTest(t)
		want := errors.New("begin failed")
		mock.ExpectBegin().WillReturnError(want)
		ctx, err := repo.BeginTx()
		if !errors.Is(err, want) {
			t.Fatalf("error = %v, want %v", err, want)
		}
		if ctx == nil || ctx.Value(authDomain.ContextTxKey) != nil || ctx.Err() != nil {
			t.Fatalf("context = %v, want valid context without transaction", ctx)
		}
	})
}

func TestBuyerInsertTxSuccess(t *testing.T) {
	for _, full := range []bool{false, true} {
		t.Run(fmt.Sprintf("full=%t", full), func(t *testing.T) {
			repo, mock := newBuyerRepoTest(t)
			repo.db.SetMaxOpenConns(1)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			mock.ExpectBegin()
			tx, err := repo.db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			buyer := sampleBuyer(full)
			mock.ExpectQuery(insertBuyerSQL).
				WithArgs(buyer.UserID, buyer.Name, buyer.Surname, buyer.BirthDate, buyer.Description, buyer.AvatarURL, buyer.Telephone).
				WillReturnRows(sqlmock.NewRows([]string{"unused"}))
			mock.ExpectRollback()
			t.Cleanup(func() {
				// InsertTx does not Scan its QueryRow. Cancellation releases the open
				// rows before rollback; this test does not require different production behavior.
				cancel()
				_ = tx.Rollback()
				deadline := time.Now().Add(5 * time.Second)
				for repo.db.Stats().InUse != 0 {
					if time.Now().After(deadline) {
						t.Error("timed out waiting for transaction cleanup")
						break
					}
					time.Sleep(time.Millisecond)
				}
			})
			if err := repo.InsertTx(context.WithValue(ctx, authDomain.ContextTxKey, tx), buyer); err != nil {
				t.Fatalf("InsertTx error = %v", err)
			}
		})
	}
}

func TestBuyerInsertTxErrors(t *testing.T) {
	duplicate := &pgconn.PgError{Code: "23505", Message: "duplicate buyer"}
	other := &pgconn.PgError{Code: "23503", Message: "foreign key violation"}
	plain := errors.New("query failed")
	for _, tc := range []struct {
		name            string
		driverErr, want error
	}{
		{"duplicate", duplicate, domain.ErrUserIDAlreadyExists},
		{"wrapped duplicate", fmt.Errorf("query: %w", duplicate), domain.ErrUserIDAlreadyExists},
		{"other postgres", other, other},
		{"plain", plain, plain},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newBuyerRepoTest(t)
			ctx := beginBuyerTestTx(t, repo, mock)
			buyer := sampleBuyer(true)
			mock.ExpectQuery(insertBuyerSQL).
				WithArgs(buyer.UserID, buyer.Name, buyer.Surname, buyer.BirthDate, buyer.Description, buyer.AvatarURL, buyer.Telephone).
				WillReturnError(tc.driverErr)
			if err := repo.InsertTx(ctx, buyer); err != tc.want {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestBuyerUpdateTx(t *testing.T) {
	execErr := errors.New("exec failed")
	pgErr := &pgconn.PgError{Code: "23505", Message: "duplicate"}
	for _, tc := range []struct {
		name   string
		full   bool
		result driver.Result
		want   error
	}{
		{"full args", true, sqlmock.NewResult(0, 1), nil},
		{"nullable args", false, sqlmock.NewResult(0, 1), nil},
		// Characterization: UpdateTx does not inspect RowsAffected or its error.
		{"zero rows ignored", true, sqlmock.NewResult(0, 0), nil},
		{"result error ignored", true, sqlmock.NewErrorResult(errors.New("result failed")), nil},
		{"exec error", true, nil, execErr},
		{"postgres error unchanged", true, nil, pgErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newBuyerRepoTest(t)
			ctx := beginBuyerTestTx(t, repo, mock)
			buyer := sampleBuyer(tc.full)
			expect := mock.ExpectExec(updateBuyerSQL).
				WithArgs(buyer.Name, buyer.Surname, buyer.BirthDate, buyer.Description, buyer.AvatarURL, buyer.Telephone, buyer.UserID)
			if tc.want != nil {
				expect.WillReturnError(tc.want)
			} else {
				expect.WillReturnResult(tc.result)
			}
			if err := repo.UpdateTx(ctx, buyer); err != tc.want {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestBuyerInvalidTransactionContext(t *testing.T) {
	for _, tc := range []struct {
		name string
		ctx  context.Context
	}{
		{"missing", context.Background()},
		{"wrong type", context.WithValue(context.Background(), authDomain.ContextTxKey, "not a transaction")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, _ := newBuyerRepoTest(t)
			want := fmt.Sprintf("'%s': %s", authDomain.ContextTxKey, authDomain.ErrContextKey)
			for name, call := range map[string]func(context.Context, domain.Buyer) error{
				"insert": repo.InsertTx, "update": repo.UpdateTx,
			} {
				if err := call(tc.ctx, sampleBuyer(true)); err == nil || err.Error() != want {
					t.Fatalf("%s error = %v, want %q", name, err, want)
				}
			}
			// Missing/invalid transactions must not contact the driver.
			repo.CommitTx(tc.ctx)
			repo.RollbackTx(tc.ctx)
		})
	}
}

func TestBuyerFinishTx(t *testing.T) {
	for _, commit := range []bool{false, true} {
		for _, driverErr := range []error{nil, errors.New("finish failed")} {
			t.Run(fmt.Sprintf("commit=%t/error=%v", commit, driverErr), func(t *testing.T) {
				repo, mock := newBuyerRepoTest(t)
				mock.ExpectBegin()
				ctx, err := repo.BeginTx()
				if err != nil {
					t.Fatal(err)
				}
				// Characterization: both void methods ignore driver errors.
				if commit {
					mock.ExpectCommit().WillReturnError(driverErr)
					repo.CommitTx(ctx)
				} else {
					mock.ExpectRollback().WillReturnError(driverErr)
					repo.RollbackTx(ctx)
				}
				if err := ctx.Value(authDomain.ContextTxKey).(*sql.Tx).Commit(); !errors.Is(err, sql.ErrTxDone) {
					t.Fatalf("Commit after finish = %v, want ErrTxDone", err)
				}
			})
		}
	}
}
