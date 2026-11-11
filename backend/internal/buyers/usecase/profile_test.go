package usecase

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	authDomain "vibe_market/backend/internal/auth/domain"
	"vibe_market/backend/internal/buyers/domain"
)

// Unconfigured methods panic so unexpected repository calls fail the test.
type profileBuyerRepoStub struct {
	selectWithUser func(int64) (domain.Buyer, authDomain.User, error)
	begin          func() (context.Context, error)
	update         func(context.Context, domain.Buyer) error
	commit         func(context.Context)
	rollback       func(context.Context)
}

var _ domain.BuyerRepo = (*profileBuyerRepoStub)(nil)

func (s *profileBuyerRepoStub) SelectByUserIDWithUser(id int64) (domain.Buyer, authDomain.User, error) {
	return s.selectWithUser(id)
}
func (s *profileBuyerRepoStub) BeginTx() (context.Context, error) { return s.begin() }
func (s *profileBuyerRepoStub) UpdateTx(ctx context.Context, buyer domain.Buyer) error {
	return s.update(ctx, buyer)
}
func (s *profileBuyerRepoStub) CommitTx(ctx context.Context)   { s.commit(ctx) }
func (s *profileBuyerRepoStub) RollbackTx(ctx context.Context) { s.rollback(ctx) }
func (*profileBuyerRepoStub) SelectByUserID(int64) (domain.Buyer, error) {
	panic("unexpected SelectByUserID")
}
func (*profileBuyerRepoStub) InsertTx(context.Context, domain.Buyer) error {
	panic("unexpected InsertTx")
}

type profileUserRepoStub struct {
	updateEmail func(context.Context, string, int64) error
}

var _ authDomain.UserRepo = (*profileUserRepoStub)(nil)

func (s *profileUserRepoStub) UpdateEmailByUserIDTx(ctx context.Context, email string, id int64) error {
	return s.updateEmail(ctx, email, id)
}
func (*profileUserRepoStub) BeginTx() (context.Context, error) { panic("unexpected user BeginTx") }
func (*profileUserRepoStub) InsertTx(context.Context, string, string) (authDomain.User, error) {
	panic("unexpected user InsertTx")
}
func (*profileUserRepoStub) CommitTx(context.Context)   { panic("unexpected user CommitTx") }
func (*profileUserRepoStub) RollbackTx(context.Context) { panic("unexpected user RollbackTx") }
func (*profileUserRepoStub) UpdatePasswordByUserID(int64, string) error {
	panic("unexpected UpdatePasswordByUserID")
}
func (*profileUserRepoStub) SelectByEmail(string) (authDomain.User, error) {
	panic("unexpected SelectByEmail")
}
func (*profileUserRepoStub) SelectByUserID(int64) (authDomain.User, error) {
	panic("unexpected user SelectByUserID")
}

func profileString(s string) *string { return &s }

func TestGetBuyerProfileUsecase(t *testing.T) {
	birth := time.Date(1995, time.March, 14, 0, 0, 0, 0, time.UTC)
	repoErr := errors.New("read failed")
	wrappedNotFound := fmt.Errorf("select: %w", authDomain.ErrUserIDNotFound)
	for _, tc := range []struct {
		name    string
		buyer   domain.Buyer
		user    authDomain.User
		err     error
		want    BuyerProfileOutDTO
		wantErr error
	}{
		{
			name: "all nullable fields populated",
			buyer: domain.Buyer{UserID: 42, Name: "Anna", Surname: profileString("Ivanova"),
				Telephone: profileString("+79991234567"), BirthDate: &birth,
				Description: profileString("About me"), AvatarURL: profileString("/avatar.png")},
			user: authDomain.User{Email: "anna@example.com"},
			want: BuyerProfileOutDTO{Name: "Anna", Surname: "Ivanova", Telephone: "+79991234567",
				Email: "anna@example.com", BirthDate: "1995-03-14 00:00:00 +0000 UTC",
				Description: "About me", AvatarURL: "/avatar.png"},
		},
		{
			name: "nil optional fields", buyer: domain.Buyer{Name: "Anna"},
			user: authDomain.User{Email: "anna@example.com"},
			want: BuyerProfileOutDTO{Name: "Anna", Email: "anna@example.com"},
		},
		{
			name: "non-nil empty strings", buyer: domain.Buyer{Surname: profileString(""),
				Telephone: profileString(""), Description: profileString(""), AvatarURL: profileString("")},
		},
		{name: "not found", err: authDomain.ErrUserIDNotFound, wantErr: authDomain.ErrUserIDNotFound},
		{name: "wrapped not found is canonicalized", err: wrappedNotFound, wantErr: authDomain.ErrUserIDNotFound},
		{name: "repository error", err: repoErr, wantErr: repoErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			repo := &profileBuyerRepoStub{selectWithUser: func(id int64) (domain.Buyer, authDomain.User, error) {
				calls++
				if id != 42 {
					t.Fatalf("user ID = %d, want 42", id)
				}
				return tc.buyer, tc.user, tc.err
			}}
			got, err := GetBuyerProfileUsecase(42, repo)
			if err != tc.wantErr || got != tc.want || calls != 1 {
				t.Fatalf("got (%+v, %v), calls %d; want (%+v, %v), calls 1", got, err, calls, tc.want, tc.wantErr)
			}
		})
	}
}

// Each test owns its fixture. The distinct context must be passed unchanged to
// both repositories, commit, and rollback, not merely carry equivalent values.
type profileTxContextKey struct{}

type profilePatchFixture struct {
	buyerRepo *profileBuyerRepoStub
	userRepo  *profileUserRepoStub
	ctx       context.Context
	calls     []string
	current   domain.Buyer
	final     domain.Buyer
	finalUser authDomain.User
	updated   domain.Buyer
	email     string
	readErr   error
	beginErr  error
	userErr   error
	buyerErr  error
	finalErr  error
}

func newProfilePatchFixture(t *testing.T) *profilePatchFixture {
	t.Helper()
	f := &profilePatchFixture{
		ctx:     context.WithValue(context.Background(), profileTxContextKey{}, new(int)),
		current: domain.Buyer{UserID: 42, AvatarURL: profileString("/old.png")},
		final: domain.Buyer{UserID: 42, Name: "Persisted name", Surname: profileString("Persisted surname"),
			Telephone: profileString("Persisted phone"), Description: profileString("Persisted description"),
			AvatarURL: profileString("/persisted.png")},
		finalUser: authDomain.User{UserID: 42, Email: "persisted@example.com"},
	}
	checkContext := func(ctx context.Context) {
		t.Helper()
		if ctx != f.ctx {
			t.Fatalf("transaction context = %v, want exact BeginTx context %v", ctx, f.ctx)
		}
	}
	reads := 0
	f.buyerRepo = &profileBuyerRepoStub{
		selectWithUser: func(id int64) (domain.Buyer, authDomain.User, error) {
			if id != 42 {
				t.Fatalf("read user ID = %d, want 42", id)
			}
			reads++
			switch reads {
			case 1:
				f.calls = append(f.calls, "read current")
				return f.current, authDomain.User{}, f.readErr
			case 2:
				f.calls = append(f.calls, "read final")
				return f.final, f.finalUser, f.finalErr
			default:
				t.Fatal("unexpected extra profile read")
				return domain.Buyer{}, authDomain.User{}, nil
			}
		},
		begin: func() (context.Context, error) {
			f.calls = append(f.calls, "begin")
			return f.ctx, f.beginErr
		},
		update: func(ctx context.Context, buyer domain.Buyer) error {
			checkContext(ctx)
			f.calls = append(f.calls, "update buyer")
			f.updated = buyer
			return f.buyerErr
		},
		commit: func(ctx context.Context) {
			checkContext(ctx)
			f.calls = append(f.calls, "commit")
		},
		rollback: func(ctx context.Context) {
			checkContext(ctx)
			f.calls = append(f.calls, "rollback")
		},
	}
	f.userRepo = &profileUserRepoStub{updateEmail: func(ctx context.Context, email string, id int64) error {
		checkContext(ctx)
		if id != 42 {
			t.Fatalf("email update user ID = %d, want 42", id)
		}
		f.calls = append(f.calls, "update email")
		f.email = email
		return f.userErr
	}}
	return f
}

func (f *profilePatchFixture) assertCalls(t *testing.T, want ...string) {
	t.Helper()
	if !reflect.DeepEqual(f.calls, want) {
		t.Fatalf("calls = %v, want %v", f.calls, want)
	}
}

func profileValidPatch() PatchBuyerProfileInDTO {
	return PatchBuyerProfileInDTO{UserID: 42, Name: "Anna", Email: "anna@example.com"}
}

func TestPatchBuyerProfileSuccess(t *testing.T) {
	birth := time.Date(2000, time.February, 29, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name      string
		dto       PatchBuyerProfileInDTO
		oldAvatar *string
		wantBuyer domain.Buyer
		wantEmail string
	}{
		{
			name: "trim all text and parse leap date",
			dto: PatchBuyerProfileInDTO{UserID: 42, Name: " \tAnna\n", Surname: " Ivanova\t",
				Telephone: " +79991234567\n", Email: "\tanna@example.com \n", BirthDate: "2000-02-29",
				Description: "\u2003About me\u2003"},
			oldAvatar: profileString("/old.png"),
			wantBuyer: domain.Buyer{UserID: 42, Name: "Anna", Surname: profileString("Ivanova"),
				Telephone: profileString("+79991234567"), BirthDate: &birth,
				Description: profileString("About me"), AvatarURL: profileString("/old.png")},
			wantEmail: "anna@example.com",
		},
		{
			name: "blank name accepted and optional fields become nil",
			dto: PatchBuyerProfileInDTO{UserID: 42, Name: " \t", Surname: "\n", Telephone: "\t ",
				Description: "\u2003", Email: "anna@example.com"},
			wantBuyer: domain.Buyer{UserID: 42}, wantEmail: "anna@example.com",
		},
		{
			name: "empty optional fields and nil avatar preserved", dto: profileValidPatch(),
			wantBuyer: domain.Buyer{UserID: 42, Name: "Anna"}, wantEmail: "anna@example.com",
		},
		{
			name: "existing avatar is preserved without normalization", dto: profileValidPatch(),
			oldAvatar: profileString("legacy.png"),
			wantBuyer: domain.Buyer{UserID: 42, Name: "Anna", AvatarURL: profileString("legacy.png")},
			wantEmail: "anna@example.com",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newProfilePatchFixture(t)
			f.current.AvatarURL = tc.oldAvatar
			got, err := PatchBuyerProfile(tc.dto, f.buyerRepo, f.userRepo)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(f.updated, tc.wantBuyer) || f.email != tc.wantEmail {
				t.Fatalf("updated = %+v, email = %q; want %+v, %q", f.updated, f.email, tc.wantBuyer, tc.wantEmail)
			}
			want := BuyerProfileOutDTO{Name: "Persisted name", Surname: "Persisted surname",
				Telephone: "Persisted phone", Email: "persisted@example.com",
				Description: "Persisted description", AvatarURL: "/persisted.png"}
			if got != want {
				t.Fatalf("output = %+v, want final read %+v", got, want)
			}
			f.assertCalls(t, "read current", "begin", "update email", "update buyer", "commit", "read final", "rollback")
		})
	}
}

func TestPatchBuyerProfileSuppliedAvatar(t *testing.T) {
	for _, tc := range []struct{ supplied, want string }{
		{"/new.png", "/new.png"},
		{"new.png", "/new.png"},
		{"", "/"},
		{"//new.png", "//new.png"},
		{" new.png ", "/ new.png "},
	} {
		t.Run(fmt.Sprintf("avatar %q", tc.supplied), func(t *testing.T) {
			f := newProfilePatchFixture(t)
			dto := profileValidPatch()
			dto.AvatarURL = profileString(tc.supplied)
			_, err := PatchBuyerProfile(dto, f.buyerRepo, f.userRepo)
			if err != nil {
				t.Fatal(err)
			}
			if f.updated.AvatarURL == nil || *f.updated.AvatarURL != tc.want {
				t.Fatalf("avatar = %v, want %q", f.updated.AvatarURL, tc.want)
			}
			if *dto.AvatarURL != tc.supplied {
				t.Fatalf("input avatar mutated: %q", *dto.AvatarURL)
			}
			f.assertCalls(t, "read current", "begin", "update email", "update buyer", "commit", "read final", "rollback")
		})
	}
}

func TestPatchBuyerProfileValidation(t *testing.T) {
	for _, tc := range []struct {
		name, email, date string
		dateError         bool
	}{
		{name: "empty email"},
		{name: "whitespace email", email: " \t\n"},
		{name: "malformed email", email: "not-an-email"},
		{name: "uppercase email is not normalized", email: "Anna@example.com"},
		{name: "email validation precedes date", email: "invalid", date: "invalid"},
		{name: "invalid date format", email: "anna@example.com", date: "29.02.2000", dateError: true},
		{name: "impossible date", email: "anna@example.com", date: "2001-02-29", dateError: true},
		{name: "date is not trimmed", email: "anna@example.com", date: " 2000-02-29 ", dateError: true},
		{name: "whitespace date", email: "anna@example.com", date: " ", dateError: true},
		{name: "datetime rejected", email: "anna@example.com", date: "2000-02-29T00:00:00Z", dateError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newProfilePatchFixture(t)
			dto := profileValidPatch()
			dto.Email, dto.BirthDate = tc.email, tc.date
			got, err := PatchBuyerProfile(dto, f.buyerRepo, f.userRepo)
			if tc.dateError {
				var parseErr *time.ParseError
				if !errors.As(err, &parseErr) || parseErr.Value != tc.date || parseErr.Layout != "2006-01-02" {
					t.Fatalf("error = %v, want date ParseError for %q", err, tc.date)
				}
			} else if err != authDomain.ErrInvalidEmail {
				t.Fatalf("error = %v, want ErrInvalidEmail", err)
			}
			if got != (BuyerProfileOutDTO{}) {
				t.Fatalf("output on validation error = %+v", got)
			}
			f.assertCalls(t)
		})
	}
}

func TestPatchBuyerProfileRepositoryErrors(t *testing.T) {
	failure := errors.New("repository failure")
	wrappedNotFound := fmt.Errorf("read: %w", authDomain.ErrUserIDNotFound)
	for _, tc := range []struct {
		name      string
		configure func(*profilePatchFixture)
		wantErr   error
		calls     []string
	}{
		{"initial read", func(f *profilePatchFixture) { f.readErr = failure }, failure,
			[]string{"read current"}},
		{"initial read preserves wrapped error", func(f *profilePatchFixture) { f.readErr = wrappedNotFound }, wrappedNotFound,
			[]string{"read current"}},
		{"begin", func(f *profilePatchFixture) { f.beginErr = failure }, failure,
			[]string{"read current", "begin"}},
		{"email update", func(f *profilePatchFixture) { f.userErr = failure }, failure,
			[]string{"read current", "begin", "update email", "rollback"}},
		{"email already exists", func(f *profilePatchFixture) { f.userErr = authDomain.ErrEmailAlreadyExists }, authDomain.ErrEmailAlreadyExists,
			[]string{"read current", "begin", "update email", "rollback"}},
		{"buyer update", func(f *profilePatchFixture) { f.buyerErr = failure }, failure,
			[]string{"read current", "begin", "update email", "update buyer", "rollback"}},
		{"final read after commit", func(f *profilePatchFixture) { f.finalErr = failure }, failure,
			[]string{"read current", "begin", "update email", "update buyer", "commit", "read final", "rollback"}},
		{"final read canonicalizes not found", func(f *profilePatchFixture) { f.finalErr = wrappedNotFound }, authDomain.ErrUserIDNotFound,
			[]string{"read current", "begin", "update email", "update buyer", "commit", "read final", "rollback"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newProfilePatchFixture(t)
			tc.configure(f)
			got, err := PatchBuyerProfile(profileValidPatch(), f.buyerRepo, f.userRepo)
			if err != tc.wantErr || got != (BuyerProfileOutDTO{}) {
				t.Fatalf("got (%+v, %v), want (zero DTO, %v)", got, err, tc.wantErr)
			}
			f.assertCalls(t, tc.calls...)
		})
	}
}

func TestPatchBuyerProfileCommitWithoutPersistence(t *testing.T) {
	// CommitTx has no error return. A silent failure cannot be signaled by the
	// BuyerRepo interface; characterize the observable result with stale data.
	f := newProfilePatchFixture(t)
	f.final = f.current
	f.final.Name = "Old name"
	f.finalUser.Email = "old@example.com"
	got, err := PatchBuyerProfile(profileValidPatch(), f.buyerRepo, f.userRepo)
	want := BuyerProfileOutDTO{Name: "Old name", Email: "old@example.com", AvatarURL: "/old.png"}
	if err != nil || got != want {
		t.Fatalf("got (%+v, %v), want stale final read (%+v, nil)", got, err, want)
	}
	f.assertCalls(t, "read current", "begin", "update email", "update buyer", "commit", "read final", "rollback")
}
