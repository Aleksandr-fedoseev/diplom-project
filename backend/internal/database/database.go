// Package database открывает пул соединений с PostgreSQL и применяет миграции.
package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Файлы миграций вшиваются в бинарник, поэтому их не нужно копировать на сервер отдельно.
//
//go:embed migrations/*.sql
var migrations embed.FS

// Connect пробует подключиться в течение 30 секунд: при старте в Docker
// PostgreSQL может быть ещё не готов, хотя контейнер уже запущен.
func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("подключение к БД: %w", err)
	}

	deadline := time.Now().Add(30 * time.Second)
	for {
		err = pool.Ping(ctx)
		if err == nil {
			return pool, nil
		}
		if time.Now().After(deadline) {
			pool.Close()
			return nil, fmt.Errorf("проверка соединения с БД: %w", err)
		}
		slog.Info("БД ещё не готова, ждём", "err", err)
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

// Migrate приводит схему БД к последней версии. goose помнит,
// какие миграции уже применены (таблица goose_db_version).
func Migrate(pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	defer func(db *sql.DB) { _ = db.Close() }(db)

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, "migrations")
}
