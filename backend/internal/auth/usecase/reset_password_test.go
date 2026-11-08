package usecase

import (
	"context"
	"errors"
	"testing"
	"vibe_market/backend/internal/auth/domain"

	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	updatedID   int64
	updatedHash string
	updateErr   error
}

func (m *mockUserRepo) BeginTx() (context.Context, error)                               { return nil, nil }
func (m *mockUserRepo) InsertTx(ctx context.Context, email, pw string) (domain.User, error) {
	return domain.User{}, nil
}
func (m *mockUserRepo) RollbackTx(ctx context.Context)                                {}
func (m *mockUserRepo) CommitTx(ctx context.Context)                                  {}
func (m *mockUserRepo) UpdateEmailByUserIDTx(ctx context.Context, email string, id int64) error {
	return nil
}
func (m *mockUserRepo) SelectByEmail(email string) (domain.User, error) {
	return domain.User{}, nil
}
func (m *mockUserRepo) SelectByUserID(id int64) (domain.User, error) {
	return domain.User{}, nil
}
func (m *mockUserRepo) UpdatePasswordByUserID(id int64, passwordHash string) error {
	m.updatedID = id
	m.updatedHash = passwordHash
	return m.updateErr
}

func TestResetPassword_Success(t *testing.T) {
	repo := &mockUserRepo{}
	newPass := "NewValidPass123!"

	out, err := ResetPassword(ResetPasswordInDTO{
		UserID:   42,
		Password: newPass,
	}, repo)

	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if out.PasswordErrMsg != "" {
		t.Fatalf("expected empty error message, got %s", out.PasswordErrMsg)
	}
	if repo.updatedID != 42 {
		t.Fatalf("expected updatedID 42, got %d", repo.updatedID)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.updatedHash), []byte(newPass)); err != nil {
		t.Fatalf("stored hash does not match new password: %v", err)
	}
}

func TestResetPassword_InvalidPassword(t *testing.T) {
	repo := &mockUserRepo{}
	shortPass := "short"

	out, err := ResetPassword(ResetPasswordInDTO{
		UserID:   42,
		Password: shortPass,
	}, repo)

	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
	if out.PasswordErrMsg == "" {
		t.Fatal("expected non-empty PasswordErrMsg")
	}
	if repo.updatedID != 0 {
		t.Fatalf("repo should not be called on invalid password")
	}
}

func TestResetPassword_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{
		updateErr: domain.ErrUserIDNotFound,
	}

	_, err := ResetPassword(ResetPasswordInDTO{
		UserID:   999,
		Password: "ValidPassword123!",
	}, repo)

	if !errors.Is(err, domain.ErrUserIDNotFound) {
		t.Fatalf("expected ErrUserIDNotFound, got %v", err)
	}
}
