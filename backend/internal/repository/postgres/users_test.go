package postgres_test

import (
	"context"
	"errors"
	"testing"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/infrastructure/postgres"
	repo "heritage-backend/internal/repository/postgres"
	"heritage-backend/internal/testdb"
)

// Тест репозитория на настоящей БД: проверяет SQL и перевод ошибок БД в ошибки домена.
func TestUsersRepository(t *testing.T) {
	pool := testdb.Connect(t)
	users := repo.NewUsers(pool)
	ctx := context.Background()
	email := testdb.Unique("repo") + "@example.ru"

	u, err := users.Create(ctx, auth.User{Email: email, Name: "Иван", Role: auth.RoleUser, PasswordHash: "h"})
	if err != nil || u.ID == 0 || u.CreatedAt.IsZero() || u.PasswordHash != "h" {
		t.Fatalf("создание: %+v, %v", u, err)
	}
	if _, err := users.Create(ctx, auth.User{Email: email, Name: "Дубль", Role: auth.RoleUser, PasswordHash: "h"}); !errors.Is(err, auth.ErrEmailTaken) {
		t.Fatalf("дубликат email: %v", err)
	}
	if got, err := users.ByEmail(ctx, email); err != nil || got.ID != u.ID {
		t.Fatalf("по email: %+v, %v", got, err)
	}
	if _, err := users.ByID(ctx, -1); !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("несуществующий: %v", err)
	}
	if got, err := users.SetRole(ctx, u.ID, auth.RoleModerator); err != nil || got.Role != auth.RoleModerator {
		t.Fatalf("роль: %+v, %v", got, err)
	}
	if _, err := users.SetBlocked(ctx, -1, true); !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("блокировка несуществующего: %v", err)
	}

	// Изменение в транзакции, которая откатилась, не сохраняется.
	tm := postgres.NewTxManager(pool)
	boom := errors.New("откат")
	_ = tm.Do(ctx, func(ctx context.Context) error {
		if _, err := users.SetBlocked(ctx, u.ID, true); err != nil {
			t.Fatal(err)
		}
		return boom
	})
	if got, _ := users.ByID(ctx, u.ID); got.IsBlocked {
		t.Fatal("блокировка из откаченной транзакции сохранилась")
	}
}
