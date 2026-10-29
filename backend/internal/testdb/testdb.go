// Package testdb подключает интеграционные тесты к настоящей PostgreSQL.
// Адрес берётся из TEST_DATABASE_URL; без него тесты пропускаются.
package testdb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/database"
)

// Connect открывает пул и применяет миграции, так что тест видит актуальную схему и стартовые данные.
func Connect(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задан")
	}
	db, err := database.Connect(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}

// Unique добавляет к строке метку времени, чтобы повторные запуски тестов не конфликтовали.
func Unique(prefix string) string {
	return prefix + "-" + time.Now().Format("150405.000000")
}
