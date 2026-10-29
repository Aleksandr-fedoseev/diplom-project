package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX — то, что умеют и пул, и транзакция. Сгенерированный sqlc код работает с любым из них.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txKey struct{}

// Conn возвращает соединение для запроса: транзакцию из ctx, если её открыл TxManager.Do,
// иначе обычный пул. Все репозитории получают соединение только через эту функцию —
// так транзакция сервиса охватывает несколько репозиториев, а сервис не знает о pgx.
func Conn(ctx context.Context, pool *pgxpool.Pool) DBTX {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}

// TxManager выполняет функцию в транзакции. Сервисы видят его через свой интерфейс
// с методом Do(ctx, fn) error и не зависят от этого пакета.
type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager { return &TxManager{pool: pool} }

// Do открывает транзакцию, кладёт её в ctx и вызывает fn. Ошибка или паника в fn —
// откат, иначе фиксация. Вложенный Do (транзакция уже в ctx) использует внешнюю
// транзакцию: зафиксирует её тот, кто открыл.
func (m *TxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("начало транзакции: %w", err)
	}
	// После Commit откат ничего не делает; при панике откатываем и пробрасываем её дальше.
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("фиксация транзакции: %w", err)
	}
	return nil
}

// InTx сообщает, идёт ли сейчас транзакция TxManager.Do. Нужно репозиториям,
// которые должны выполняться только внутри транзакции (например, outbox).
func InTx(ctx context.Context) bool {
	_, ok := ctx.Value(txKey{}).(pgx.Tx)
	return ok
}
