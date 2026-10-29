package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/errs"
)

// Тесты сервиса без БД и HTTP: репозиторий, хэширование и токены — фейки в памяти.

type fakeUsers struct {
	byID   map[int64]auth.User
	nextID int64
}

func newFakeUsers() *fakeUsers { return &fakeUsers{byID: map[int64]auth.User{}} }

func (f *fakeUsers) Create(_ context.Context, u auth.User) (auth.User, error) {
	for _, x := range f.byID {
		if x.Email == u.Email {
			return auth.User{}, auth.ErrEmailTaken
		}
	}
	f.nextID++
	u.ID = f.nextID
	f.byID[u.ID] = u
	return u, nil
}

func (f *fakeUsers) ByID(_ context.Context, id int64) (auth.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return auth.User{}, auth.ErrUserNotFound
	}
	return u, nil
}

func (f *fakeUsers) ByEmail(_ context.Context, email string) (auth.User, error) {
	for _, u := range f.byID {
		if u.Email == email {
			return u, nil
		}
	}
	return auth.User{}, auth.ErrUserNotFound
}

func (f *fakeUsers) List(context.Context, int, int) ([]auth.User, error) { return nil, nil }

func (f *fakeUsers) SetRole(ctx context.Context, id int64, role auth.Role) (auth.User, error) {
	u, err := f.ByID(ctx, id)
	u.Role = role
	f.byID[id] = u
	return u, err
}

func (f *fakeUsers) SetBlocked(ctx context.Context, id int64, blocked bool) (auth.User, error) {
	u, err := f.ByID(ctx, id)
	u.IsBlocked = blocked
	f.byID[id] = u
	return u, err
}

// fakeHasher «хэширует» приставкой — достаточно, чтобы проверить, что сервис хранит не пароль.
type fakeHasher struct{}

func (fakeHasher) Hash(p string) (string, error) { return "hash:" + p, nil }
func (fakeHasher) Compare(h, p string) error {
	if h != "hash:"+p {
		return errors.New("не подходит")
	}
	return nil
}

// fakeTokens — токен вида «id:role».
type fakeTokens struct{}

func (fakeTokens) Issue(id int64, role string) (string, error) {
	return fmt.Sprintf("%d:%s", id, role), nil
}
func (fakeTokens) Parse(t string) (int64, string, error) {
	var id int64
	idPart, role, ok := strings.Cut(t, ":")
	if _, err := fmt.Sscan(idPart, &id); err != nil || !ok {
		return 0, "", errors.New("плохой токен")
	}
	return id, role, nil
}

func newTestService() (*Service, *fakeUsers) {
	users := newFakeUsers()
	return NewService(users, fakeHasher{}, fakeTokens{}), users
}

func TestRegisterAndLogin(t *testing.T) {
	svc, users := newTestService()
	ctx := context.Background()

	u, err := svc.Register(ctx, auth.Registration{Email: " Ivan@Test.ru ", Password: "password1", Name: "Иван"})
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "ivan@test.ru" || u.Role != auth.RoleUser || users.byID[u.ID].PasswordHash != "hash:password1" {
		t.Fatalf("зарегистрирован: %+v", users.byID[u.ID])
	}
	if _, err := svc.Register(ctx, auth.Registration{Email: "ivan@test.ru", Password: "password1", Name: "Иван"}); !errors.Is(err, auth.ErrEmailTaken) {
		t.Fatalf("повторный email: %v", err)
	}
	if _, err := svc.Register(ctx, auth.Registration{Email: "x@y.ru", Password: "short", Name: "Иван"}); !errs.Is(err, errs.Invalid) {
		t.Fatalf("короткий пароль: %v", err)
	}

	// Неверный пароль и неизвестный email — одна и та же ошибка.
	if _, _, err := svc.Login(ctx, "ivan@test.ru", "wrong"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("неверный пароль: %v", err)
	}
	if _, _, err := svc.Login(ctx, "nobody@test.ru", "password1"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("неизвестный email: %v", err)
	}

	token, logged, err := svc.Login(ctx, "IVAN@test.ru", "password1")
	if err != nil || logged.ID != u.ID {
		t.Fatalf("вход: %v, %+v", err, logged)
	}
	p, err := svc.Authenticate(ctx, token)
	if err != nil || p != (auth.Principal{UserID: u.ID, Role: auth.RoleUser}) {
		t.Fatalf("Authenticate: %+v, %v", p, err)
	}
	if _, err := svc.Authenticate(ctx, "мусор"); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("плохой токен: %v", err)
	}
	if _, err := svc.Authenticate(ctx, "1:superuser"); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("токен с неизвестной ролью: %v", err)
	}
}

func TestBlockedUser(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	if err := svc.EnsureAdmin(ctx, "admin@test.ru", "adminpass1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.EnsureAdmin(ctx, "admin@test.ru", "adminpass1"); err != nil {
		t.Fatalf("повторный EnsureAdmin должен молча пропускаться: %v", err)
	}
	_, admin, _ := svc.Login(ctx, "admin@test.ru", "adminpass1")
	u, _ := svc.Register(ctx, auth.Registration{Email: "u@test.ru", Password: "password1", Name: "U"})
	actor := auth.Principal{UserID: admin.ID, Role: admin.Role}

	if _, err := svc.SetBlocked(ctx, actor, u.ID, true); err != nil {
		t.Fatal(err)
	}
	if err := svc.EnsureActive(ctx, u.ID); !errors.Is(err, auth.ErrBlocked) {
		t.Fatalf("EnsureActive заблокированного: %v", err)
	}
	if _, _, err := svc.Login(ctx, "u@test.ru", "password1"); !errors.Is(err, auth.ErrBlocked) {
		t.Fatalf("вход заблокированного: %v", err)
	}
	if _, err := svc.SetBlocked(ctx, actor, 999, true); !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("блокировка несуществующего: %v", err)
	}
}

func TestSetRoleAndModeratorLimits(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	admin := auth.Principal{UserID: 100, Role: auth.RoleAdmin}
	a, _ := svc.Register(ctx, auth.Registration{Email: "a@test.ru", Password: "password1", Name: "A"})
	b, _ := svc.Register(ctx, auth.Registration{Email: "b@test.ru", Password: "password1", Name: "B"})

	mod, err := svc.SetRole(ctx, admin, a.ID, auth.RoleModerator)
	if err != nil || mod.Role != auth.RoleModerator {
		t.Fatalf("назначение модератора: %+v, %v", mod, err)
	}
	if _, err := svc.SetRole(ctx, admin, b.ID, auth.Role("root")); !errors.Is(err, auth.ErrInvalidRole) {
		t.Fatalf("неизвестная роль: %v", err)
	}

	// Модератор может блокировать пользователя, но не другого модератора.
	modActor := auth.Principal{UserID: a.ID, Role: auth.RoleModerator}
	_, _ = svc.SetRole(ctx, admin, b.ID, auth.RoleModerator)
	if _, err := svc.SetBlocked(ctx, modActor, b.ID, true); !errors.Is(err, auth.ErrCannotBlockStaff) {
		t.Fatalf("модератор блокирует модератора: %v", err)
	}
}
