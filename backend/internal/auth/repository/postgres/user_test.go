package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	"vibe_market/backend/internal/auth/domain"
)

const (
	insertUserSQL     = `INSERT INTO "user" (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash, created_at, updated_at;`
	selectEmailSQL    = `SELECT id, email, password_hash, created_at, updated_at FROM "user" WHERE email = $1;`
	selectIDSQL       = `SELECT id, email, password_hash, created_at, updated_at FROM "user" WHERE id = $1;`
	updateEmailSQL    = `UPDATE "user" SET email = $1 WHERE id = $2`
	updatePasswordSQL = `UPDATE "user" SET password_hash = $1 WHERE id = $2;`
)

func newUserRepoTest(t *testing.T) (*UserRepoPostgres, sqlmock.Sqlmock) {
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
	return NewUserRepo(db), mock
}

func userRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"})
}

func sampleUser() domain.User {
	return domain.User{
		UserID: 42, Email: "user@example.com", PasswordHash: "password-hash",
		CreatedAt: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC),
	}
}

func rowsForUser(user domain.User) *sqlmock.Rows {
	return userRows().AddRow(user.UserID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
}

func beginUserTestTx(t *testing.T, repo *UserRepoPostgres, mock sqlmock.Sqlmock) context.Context {
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

func TestNewUserRepo(t *testing.T) {
	db := &sql.DB{}
	repo := NewUserRepo(db)
	if repo == nil || repo.db != db {
		t.Fatalf("NewUserRepo() = %#v, want repository retaining supplied DB", repo)
	}
}

func TestUserBeginTx(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newUserRepoTest(t)
		ctx := beginUserTestTx(t, repo, mock)
		if tx, ok := ctx.Value(domain.ContextTxKey).(*sql.Tx); !ok || tx == nil {
			t.Fatalf("transaction in context = %v, want *sql.Tx", ctx.Value(domain.ContextTxKey))
		}
		if ctx.Err() != nil {
			t.Fatalf("context error = %v", ctx.Err())
		}
	})
	t.Run("begin error", func(t *testing.T) {
		repo, mock := newUserRepoTest(t)
		want := errors.New("begin failed")
		mock.ExpectBegin().WillReturnError(want)
		ctx, err := repo.BeginTx()
		if !errors.Is(err, want) {
			t.Fatalf("error = %v, want %v", err, want)
		}
		if ctx == nil || ctx.Value(domain.ContextTxKey) != nil || ctx.Err() != nil {
			t.Fatalf("context = %v, want valid context without transaction", ctx)
		}
	})
}

func TestUserFinishTx(t *testing.T) {
	for _, commit := range []bool{false, true} {
		for _, driverErr := range []error{nil, errors.New("transaction finish failed")} {
			t.Run(fmt.Sprintf("commit=%t/error=%v", commit, driverErr), func(t *testing.T) {
				repo, mock := newUserRepoTest(t)
				mock.ExpectBegin()
				ctx, err := repo.BeginTx()
				if err != nil {
					t.Fatal(err)
				}
				if commit {
					mock.ExpectCommit().WillReturnError(driverErr)
					repo.CommitTx(ctx)
				} else {
					mock.ExpectRollback().WillReturnError(driverErr)
					repo.RollbackTx(ctx)
				}
				if err := ctx.Value(domain.ContextTxKey).(*sql.Tx).Commit(); !errors.Is(err, sql.ErrTxDone) {
					t.Fatalf("Commit after finish = %v, want ErrTxDone", err)
				}
			})
		}
	}
}

func TestUserInvalidTransactionContext(t *testing.T) {
	for _, tc := range []struct {
		name string
		ctx  context.Context
	}{
		{"missing", context.Background()},
		{"wrong type", context.WithValue(context.Background(), domain.ContextTxKey, "not a transaction")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, _ := newUserRepoTest(t)
			want := fmt.Sprintf("'%s': %s", domain.ContextTxKey, domain.ErrContextKey)
			user, err := repo.InsertTx(tc.ctx, "email", "hash")
			if err == nil || err.Error() != want || user != (domain.User{}) {
				t.Fatalf("InsertTx = (%#v, %v), want zero user and %q", user, err, want)
			}
			if err := repo.UpdateEmailByUserIDTx(tc.ctx, "email", 42); err == nil || err.Error() != want {
				t.Fatalf("UpdateEmailByUserIDTx error = %v, want %q", err, want)
			}
			// Neither method should contact the driver without a transaction.
			repo.CommitTx(tc.ctx)
			repo.RollbackTx(tc.ctx)
		})
	}
}

// QueryRow's Scan paths are shared by insertion and both lookup methods.
func TestUserReturningQueries(t *testing.T) {
	for _, method := range []string{"insert", "select email", "select ID"} {
		t.Run(method, func(t *testing.T) {
			queryErr := errors.New("query failed")
			fetchErr := errors.New("row fetch failed")
			want := sampleUser()
			noRowsErr := sql.ErrNoRows
			if method == "select email" {
				noRowsErr = domain.ErrEmailNotFound
			} else if method == "select ID" {
				noRowsErr = domain.ErrUserIDNotFound
			}
			for _, tc := range []struct {
				name     string
				rows     func() *sqlmock.Rows
				queryErr error
				wantErr  error
				scanErr  bool
				wantUser domain.User
			}{
				{name: "success mapping", rows: func() *sqlmock.Rows { return rowsForUser(want) }, wantUser: want},
				{name: "query error", queryErr: queryErr, wantErr: queryErr},
				{name: "no rows", rows: userRows, wantErr: noRowsErr},
				{name: "scan error", rows: func() *sqlmock.Rows {
					return userRows().AddRow("invalid ID", want.Email, want.PasswordHash, want.CreatedAt, want.UpdatedAt)
				}, scanErr: true},
				{name: "row fetch error", rows: func() *sqlmock.Rows { return rowsForUser(want).RowError(0, fetchErr) }, wantErr: fetchErr},
			} {
				t.Run(tc.name, func(t *testing.T) {
					repo, mock := newUserRepoTest(t)
					var call func() (domain.User, error)
					var query string
					var args []driver.Value
					switch method {
					case "insert":
						ctx := beginUserTestTx(t, repo, mock)
						query, args = insertUserSQL, []driver.Value{want.Email, want.PasswordHash}
						call = func() (domain.User, error) { return repo.InsertTx(ctx, want.Email, want.PasswordHash) }
					case "select email":
						query, args = selectEmailSQL, []driver.Value{want.Email}
						call = func() (domain.User, error) { return repo.SelectByEmail(want.Email) }
					case "select ID":
						query, args = selectIDSQL, []driver.Value{want.UserID}
						call = func() (domain.User, error) { return repo.SelectByUserID(want.UserID) }
					}
					expect := mock.ExpectQuery(query).WithArgs(args...)
					if tc.queryErr != nil {
						expect.WillReturnError(tc.queryErr)
					} else {
						expect.WillReturnRows(tc.rows()).RowsWillBeClosed()
					}
					got, err := call()
					if tc.scanErr {
						if err == nil || !strings.Contains(err.Error(), `column index 0, name "id"`) {
							t.Fatalf("error = %v, want ID scan conversion error", err)
						}
					} else if !errors.Is(err, tc.wantErr) {
						t.Fatalf("error = %v, want %v", err, tc.wantErr)
					}
					if got != tc.wantUser {
						t.Fatalf("user = %#v, want %#v", got, tc.wantUser)
					}
				})
			}
		})
	}
}

func TestUserTransactionalQueryErrors(t *testing.T) {
	duplicate := &pgconn.PgError{Code: "23505", Message: "duplicate email"}
	other := &pgconn.PgError{Code: "23503", Message: "foreign key violation"}
	plain := errors.New("query failed")
	for _, method := range []string{"insert", "update email"} {
		for _, tc := range []struct {
			name      string
			driverErr error
			want      error
		}{
			{"duplicate direct", duplicate, domain.ErrEmailAlreadyExists},
			{"duplicate wrapped", fmt.Errorf("query: %w", duplicate), domain.ErrEmailAlreadyExists},
			{"other postgres", other, other},
			{"other postgres wrapped", fmt.Errorf("query: %w", other), other},
			{"plain error", plain, plain},
		} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				repo, mock := newUserRepoTest(t)
				ctx := beginUserTestTx(t, repo, mock)
				var err error
				if method == "insert" {
					mock.ExpectQuery(insertUserSQL).WithArgs("email", "hash").WillReturnError(tc.driverErr)
					var user domain.User
					user, err = repo.InsertTx(ctx, "email", "hash")
					if user != (domain.User{}) {
						t.Fatalf("user = %#v, want zero user", user)
					}
				} else {
					mock.ExpectQuery(updateEmailSQL).WithArgs("email", int64(42)).WillReturnError(tc.driverErr)
					err = repo.UpdateEmailByUserIDTx(ctx, "email", 42)
				}
				if !errors.Is(err, tc.want) {
					t.Fatalf("error = %v, want %v", err, tc.want)
				}
				if tc.want != domain.ErrEmailAlreadyExists && err != tc.driverErr {
					t.Fatalf("error = %v, want original driver error %v", err, tc.driverErr)
				}
			})
		}
	}
}

func TestUserUpdateEmailSuccess(t *testing.T) {
	repo, mock := newUserRepoTest(t)
	repo.db.SetMaxOpenConns(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mock.ExpectBegin()
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	txCtx := context.WithValue(ctx, domain.ContextTxKey, tx)
	mock.ExpectQuery(updateEmailSQL).WithArgs("new@example.com", int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"email"}))
	mock.ExpectRollback()
	t.Cleanup(func() {
		// Production calls QueryRow without Scan, leaving its rows open. Cancel
		// the transaction context to release them; do not assert that this leak
		// is desirable behavior or require RowsWillBeClosed from this method.
		cancel()
		_ = tx.Rollback()
		// Automatic rollback may win the race with Rollback and discard the
		// connection. Wait for its release before checking driver expectations.
		deadline := time.Now().Add(5 * time.Second)
		for repo.db.Stats().InUse != 0 {
			if time.Now().After(deadline) {
				t.Error("timed out waiting for transaction cleanup")
				break
			}
			time.Sleep(time.Millisecond)
		}
	})
	if err := repo.UpdateEmailByUserIDTx(txCtx, "new@example.com", 42); err != nil {
		t.Fatalf("UpdateEmailByUserIDTx error = %v", err)
	}
}

func TestUserUpdatePasswordByUserID(t *testing.T) {
	queryErr := errors.New("exec failed")
	resultErr := errors.New("rows affected failed")
	for _, tc := range []struct {
		name    string
		result  driver.Result
		execErr error
		want    error
	}{
		{"one affected", sqlmock.NewResult(0, 1), nil, nil},
		{"zero affected", sqlmock.NewResult(0, 0), nil, domain.ErrUserIDNotFound},
		{"rows affected error", sqlmock.NewErrorResult(resultErr), nil, resultErr},
		{"exec error", nil, queryErr, queryErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newUserRepoTest(t)
			expect := mock.ExpectExec(updatePasswordSQL).WithArgs("new-password-hash", int64(42))
			if tc.execErr != nil {
				expect.WillReturnError(tc.execErr)
			} else {
				expect.WillReturnResult(tc.result)
			}
			if err := repo.UpdatePasswordByUserID(42, "new-password-hash"); !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}
