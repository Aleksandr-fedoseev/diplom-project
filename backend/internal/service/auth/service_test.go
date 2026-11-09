package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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

func (f *fakeUsers) UpdateName(ctx context.Context, id int64, name string) (auth.User, error) {
	u, err := f.ByID(ctx, id)
	u.Name = name
	f.byID[id] = u
	return u, err
}

func (f *fakeUsers) UpdatePassword(ctx context.Context, id int64, hash string) error {
	u, err := f.ByID(ctx, id)
	u.PasswordHash = hash
	f.byID[id] = u
	return err
}

func (f *fakeUsers) SetAvatar(ctx context.Context, id int64, key *string) (auth.User, error) {
	u, err := f.ByID(ctx, id)
	u.AvatarKey = key
	f.byID[id] = u
	return u, err
}

// fakeAvatars — хранилище файлов в памяти.
type fakeAvatars struct{ files map[string][]byte }

func newFakeAvatars() *fakeAvatars { return &fakeAvatars{files: map[string][]byte{}} }

func (f *fakeAvatars) Put(_ context.Context, key string, body io.Reader, _ int64, _ string) error {
	b, err := io.ReadAll(body)
	f.files[key] = b
	return err
}

func (f *fakeAvatars) Get(_ context.Context, key string) (io.ReadSeekCloser, error) {
	b, ok := f.files[key]
	if !ok {
		return nil, errors.New("нет файла")
	}
	return nopCloser{bytes.NewReader(b)}, nil
}

func (f *fakeAvatars) Remove(_ context.Context, key string) error {
	delete(f.files, key)
	return nil
}

type nopCloser struct{ io.ReadSeeker }

func (nopCloser) Close() error { return nil }

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
	return NewService(users, fakeHasher{}, fakeTokens{}, newFakeAvatars()), users
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

// Выданный токен не продлевает прав: снятие роли и блокировка действуют сразу.
func TestAuthenticateUsesCurrentRoleAndBlock(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	admin := auth.Principal{UserID: 100, Role: auth.RoleAdmin}
	u, _ := svc.Register(ctx, auth.Registration{Email: "m@test.ru", Password: "password1", Name: "M"})
	if _, err := svc.SetRole(ctx, admin, u.ID, auth.RoleModerator); err != nil {
		t.Fatal(err)
	}
	token, _, err := svc.Login(ctx, "m@test.ru", "password1")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.SetRole(ctx, admin, u.ID, auth.RoleUser); err != nil {
		t.Fatal(err)
	}
	if p, err := svc.Authenticate(ctx, token); err != nil || p.Role != auth.RoleUser {
		t.Fatalf("после снятия роли модератора: %+v, %v", p, err)
	}

	if _, err := svc.SetBlocked(ctx, admin, u.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Authenticate(ctx, token); !errors.Is(err, auth.ErrBlocked) {
		t.Fatalf("токен заблокированного: %v", err)
	}

	if _, err := svc.Authenticate(ctx, "999:user"); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("токен несуществующего пользователя: %v", err)
	}
}

func TestUpdateName(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	u, _ := svc.Register(ctx, auth.Registration{Email: "a@t.ru", Password: "password1", Name: "Иван"})

	got, err := svc.UpdateName(ctx, u.ID, "  Иван Петров  ")
	if err != nil || got.Name != "Иван Петров" {
		t.Fatalf("имя: %+v, %v", got, err)
	}
	for _, bad := range []string{"   ", strings.Repeat("я", auth.MaxNameLen+1)} {
		if _, err := svc.UpdateName(ctx, u.ID, bad); !errs.Is(err, errs.Invalid) {
			t.Fatalf("имя %q: %v", bad, err)
		}
	}
}

func TestChangePassword(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	u, _ := svc.Register(ctx, auth.Registration{Email: "a@t.ru", Password: "password1", Name: "Иван"})

	if err := svc.ChangePassword(ctx, u.ID, "не тот", "password2"); !errors.Is(err, auth.ErrWrongPassword) {
		t.Fatalf("неверный текущий пароль: %v", err)
	}
	if err := svc.ChangePassword(ctx, u.ID, "password1", "short"); !errs.Is(err, errs.Invalid) {
		t.Fatalf("короткий новый пароль: %v", err)
	}
	if err := svc.ChangePassword(ctx, u.ID, "password1", "password1"); !errors.Is(err, auth.ErrSamePassword) {
		t.Fatalf("тот же пароль: %v", err)
	}
	if err := svc.ChangePassword(ctx, u.ID, "password1", "password2"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Login(ctx, "a@t.ru", "password1"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("старый пароль подошёл: %v", err)
	}
	if _, _, err := svc.Login(ctx, "a@t.ru", "password2"); err != nil {
		t.Fatalf("новый пароль: %v", err)
	}
}

// Минимальный PNG: сигнатура файла — по ней определяется тип.
var pngBytes = append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 64)...)

func TestAvatar(t *testing.T) {
	users := newFakeUsers()
	store := newFakeAvatars()
	svc := NewService(users, fakeHasher{}, fakeTokens{}, store)
	ctx := context.Background()
	u, _ := svc.Register(ctx, auth.Registration{Email: "a@t.ru", Password: "password1", Name: "Иван"})

	if _, _, err := svc.OpenAvatar(ctx, u.ID); !errors.Is(err, auth.ErrNoAvatar) {
		t.Fatalf("аватара ещё нет: %v", err)
	}
	if _, err := svc.SetAvatar(ctx, u.ID, strings.NewReader("<html>не картинка</html>"), 24); !errors.Is(err, auth.ErrAvatarType) {
		t.Fatalf("не картинка: %v", err)
	}
	if _, err := svc.SetAvatar(ctx, u.ID, bytes.NewReader(pngBytes), auth.MaxAvatarBytes+1); !errors.Is(err, auth.ErrAvatarTooBig) {
		t.Fatalf("большой файл: %v", err)
	}

	first, err := svc.SetAvatar(ctx, u.ID, bytes.NewReader(pngBytes), int64(len(pngBytes)))
	if err != nil || first.AvatarKey == nil || !strings.HasSuffix(*first.AvatarKey, ".png") || first.AvatarURL() == "" {
		t.Fatalf("загрузка: %+v, %v", first, err)
	}
	body, ct, err := svc.OpenAvatar(ctx, u.ID)
	if err != nil || ct != "image/png" {
		t.Fatalf("открыть: %q, %v", ct, err)
	}
	if got, _ := io.ReadAll(body); !bytes.Equal(got, pngBytes) {
		t.Fatal("файл сохранился не целиком")
	}

	// Новый аватар заменяет старый: старый файл удаляется, адрес меняется.
	second, _ := svc.SetAvatar(ctx, u.ID, bytes.NewReader(pngBytes), int64(len(pngBytes)))
	if _, ok := store.files[*first.AvatarKey]; ok || second.AvatarURL() == first.AvatarURL() || len(store.files) != 1 {
		t.Fatalf("замена: файлы %v", store.files)
	}

	if u, err := svc.RemoveAvatar(ctx, u.ID); err != nil || u.AvatarKey != nil || len(store.files) != 0 {
		t.Fatalf("удаление: %+v, %v, %v", u, err, store.files)
	}
}
