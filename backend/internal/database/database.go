// Package database открывает пул соединений с PostgreSQL и применяет миграции.
package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Файлы миграций вшиваются в бинарник, поэтому их не нужно копировать на сервер отдельно.
//
//go:embed migrations/*.sql
var migrations embed.FS

func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("подключение к БД: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("проверка соединения с БД: %w", err)
	}
	return pool, nil
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
