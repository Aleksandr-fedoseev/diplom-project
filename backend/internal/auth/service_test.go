package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"heritage-backend/internal/httpx"
	"heritage-backend/internal/testdb"
)

// Тест работает с настоящей БД. Запуск:
//
//	TEST_DATABASE_URL=postgres://heritage:...@localhost:5433/heritage?sslmode=disable go test ./internal/auth/
func newTestService(t *testing.T) *Service {
	return NewService(testdb.Connect(t), NewTokenIssuer("0123456789abcdef0123456789abcdef", time.Hour))
}

func wantStatus(t *testing.T, err error, status int) {
	t.Helper()
	var he *httpx.Error
	if !errors.As(err, &he) || he.Status != status {
		t.Fatalf("ожидалась ошибка %d, получено: %v", status, err)
	}
}

func TestRegisterLoginGetByID(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	email := testdb.Unique("test") + "@example.ru"

	u, err := svc.Register(ctx, RegisterInput{Email: "  " + email + " ", Password: "password1", Name: "Иван"})
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != email || u.Role != RoleUser || u.ID == 0 {
		t.Fatalf("неожиданный пользователь: %+v", u)
	}

	_, err = svc.Register(ctx, RegisterInput{Email: email, Password: "password1", Name: "Иван"})
	wantStatus(t, err, 409)

	_, _, err = svc.Login(ctx, email, "wrong-password")
	wantStatus(t, err, 401)

	token, logged, err := svc.Login(ctx, email, "password1")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || logged.ID != u.ID {
		t.Fatalf("вход вернул %q, %+v", token, logged)
	}

	got, err := svc.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got != u {
		t.Fatalf("GetByID: %+v, ожидалось %+v", got, u)
	}

	_, err = svc.GetByID(ctx, -1)
	wantStatus(t, err, 404)
}
