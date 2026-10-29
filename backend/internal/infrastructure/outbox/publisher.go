// Package outbox — надёжная доставка событий между модулями (шаблон «transactional outbox»).
//
// Модуль-производитель записывает событие в таблицу outbox_events В ТОЙ ЖЕ транзакции,
// что и изменение данных (Publisher.Publish внутри TxManager.Do). Поэтому событие появится,
// только если изменение сохранилось, и не потеряется, если приложение упадёт сразу после фиксации.
// Фоновый Processor читает таблицу и вызывает подписчиков (например, модуль уведомлений).
//
// Доставка «хотя бы один раз» (at-least-once): если подписчик выполнил работу, но приложение
// упало до отметки события, событие обработается повторно. Подписчики должны это переносить
// (повторное письмо допустимо, двойное списание денег — нет).
//
// О домене пакет ничего не знает: событие — это имя и JSON.
package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/infrastructure/postgres"
	"heritage-backend/internal/repository/postgres/sqlc"
)

// ErrNoTransaction — Publish вызван вне TxManager.Do. Событие без транзакции
// могло бы уйти без изменения (или изменение без события) — ради этой связки outbox и нужен.
var ErrNoTransaction = errors.New("outbox: событие можно опубликовать только внутри транзакции")

// Publisher записывает события в outbox. Модули видят его через свой порт EventPublisher.
type Publisher struct {
	pool *pgxpool.Pool
}

func NewPublisher(pool *pgxpool.Pool) *Publisher { return &Publisher{pool: pool} }

// Publish кодирует payload в JSON и записывает событие в транзакции из ctx.
// Вне транзакции возвращает ErrNoTransaction: сначала изменение, потом отдельной
// записью событие — это ровно та несогласованность, от которой защищает outbox.
func (p *Publisher) Publish(ctx context.Context, eventType string, payload any) error {
	if !postgres.InTx(ctx) {
		return ErrNoTransaction
	}
	if eventType == "" {
		return errors.New("outbox: пустое имя события")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("outbox: кодирование события %s: %w", eventType, err)
	}
	err = sqlc.New(postgres.Conn(ctx, p.pool)).InsertOutboxEvent(ctx, sqlc.InsertOutboxEventParams{
		EventType: eventType, Payload: data,
	})
	if err != nil {
		return fmt.Errorf("outbox: запись события %s: %w", eventType, err)
	}
	return nil
}
