package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"heritage-backend/internal/infrastructure/postgres"
)

func TestTxManagerCommitRollbackNested(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задан")
	}
	ctx := context.Background()
	pool, err := postgres.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	_, _ = pool.Exec(ctx, `DROP TABLE IF EXISTS tx_probe_t`) // остаток упавшего прошлого запуска

	tm := postgres.NewTxManager(pool)
	count := func() int {
		var n int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM pg_tables WHERE tablename = 'tx_probe_t'`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	// Откат: таблица, созданная в транзакции с ошибкой, не появляется.
	boom := errors.New("ошибка в сервисе")
	err = tm.Do(ctx, func(ctx context.Context) error {
		if !postgres.InTx(ctx) {
			t.Fatal("внутри Do транзакция должна быть в ctx")
		}
		if _, err := postgres.Conn(ctx, pool).Exec(ctx, `CREATE TABLE tx_probe_t (v int)`); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) || count() != 0 {
		t.Fatalf("откат не сработал: err=%v, таблиц=%d", err, count())
	}

	// Вложенный Do работает в той же транзакции; фиксирует внешний.
	err = tm.Do(ctx, func(ctx context.Context) error {
		return tm.Do(ctx, func(ctx context.Context) error {
			_, err := postgres.Conn(ctx, pool).Exec(ctx, `CREATE TABLE tx_probe_t (v int)`)
			return err
		})
	})
	if err != nil || count() != 1 {
		t.Fatalf("фиксация вложенной транзакции: err=%v, таблиц=%d", err, count())
	}
	if _, err := pool.Exec(ctx, `DROP TABLE tx_probe_t`); err != nil {
		t.Fatal(err)
	}
	if postgres.InTx(ctx) {
		t.Fatal("вне Do транзакции быть не должно")
	}
}
