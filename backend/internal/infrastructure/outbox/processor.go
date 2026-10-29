package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/repository/postgres/sqlc"
)

const (
	maxAttempts  = 5               // после стольких неудач событие помечается failed
	batchSize    = 20              // событий за один проход
	pollInterval = 5 * time.Second // как часто Run проверяет таблицу
	// Аренда: взятое событие откладывается на это время. Если обработчик упадёт посреди
	// работы, событие само вернётся в очередь, когда аренда истечёт.
	leaseSeconds = 5 * 60
)

// Handler обрабатывает payload события (JSON). Ошибка — событие повторится позже.
// Обработчик может быть вызван повторно для того же события (доставка at-least-once).
type Handler = func(ctx context.Context, payload []byte) error

// Processor доставляет события из outbox подписчикам.
type Processor struct {
	q        *sqlc.Queries
	handlers map[string]Handler
}

func NewProcessor(pool *pgxpool.Pool) *Processor {
	return &Processor{q: sqlc.New(pool), handlers: map[string]Handler{}}
}

// Subscribe назначает обработчик типу события. Вызывается при сборке приложения,
// до запуска Run (карта обработчиков не защищена от одновременного доступа).
// На один тип — один обработчик: повторная подписка — ошибка сборки, поэтому паника.
func (p *Processor) Subscribe(eventType string, h func(ctx context.Context, payload []byte) error) {
	if _, ok := p.handlers[eventType]; ok {
		panic("outbox: на событие " + eventType + " уже есть подписчик")
	}
	p.handlers[eventType] = h
}

// Run обрабатывает события, пока не отменён ctx. Запускается в отдельной горутине.
func (p *Processor) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		if err := p.ProcessDue(ctx); err != nil && ctx.Err() == nil {
			slog.Error("обработка outbox", "err", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// ProcessDue обрабатывает одну пачку событий, срок которых наступил.
// Вынесено из Run, чтобы вызывать в тестах.
func (p *Processor) ProcessDue(ctx context.Context) error {
	events, err := p.q.ClaimDueOutboxEvents(ctx, sqlc.ClaimDueOutboxEventsParams{
		LeaseSeconds: leaseSeconds, BatchSize: batchSize,
	})
	if err != nil {
		return err
	}
	// UPDATE ... RETURNING не сохраняет порядок — обрабатываем в порядке записи,
	// чтобы «заказ оформлен» не пришёл после «заказ отменён».
	slices.SortFunc(events, func(a, b sqlc.ClaimDueOutboxEventsRow) int { return int(a.ID - b.ID) })
	for _, ev := range events {
		h, ok := p.handlers[ev.EventType]
		if !ok {
			// Никто не подписан — делать нечего, но в лог запишем: возможно, забыли подписку.
			slog.Info("у события outbox нет подписчиков", "id", ev.ID, "type", ev.EventType)
			err = p.q.MarkOutboxEventDone(ctx, ev.ID)
		} else if herr := call(ctx, h, ev.Payload); herr == nil {
			err = p.q.MarkOutboxEventDone(ctx, ev.ID)
		} else {
			slog.Warn("событие outbox не обработано", "id", ev.ID, "type", ev.EventType,
				"attempt", ev.Attempts, "err", herr)
			err = p.q.MarkOutboxEventFailed(ctx, failure(ev.ID, ev.Attempts, herr))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// call вызывает обработчик; паника в обработчике считается ошибкой, а не роняет приложение.
func call(ctx context.Context, h Handler, payload []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("паника в обработчике: %v", r)
		}
	}()
	return h(ctx, payload)
}

// failure решает, что делать с необработанным событием: повторить позже
// (через 1, 4, 9, 16 минут) или после последней попытки пометить как failed.
func failure(id int64, attempts int32, herr error) sqlc.MarkOutboxEventFailedParams {
	p := sqlc.MarkOutboxEventFailedParams{
		ID:            id,
		Status:        "pending",
		LastError:     herr.Error(),
		NextAttemptAt: time.Now().Add(time.Duration(attempts*attempts) * time.Minute),
	}
	if attempts >= maxAttempts {
		p.Status = "failed"
	}
	return p
}
