package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/infrastructure/postgres"
	"heritage-backend/internal/testdb"
)

// Интеграционные тесты на настоящей PostgreSQL. Каждый тест работает со своим
// уникальным типом события: в таблице могут лежать события прошлых запусков.

type event struct {
	ID       int64
	Status   string
	Attempts int
	Next     time.Time
	LastErr  string
}

func eventsOf(t *testing.T, db *pgxpool.Pool, eventType string) []event {
	t.Helper()
	rows, err := db.Query(context.Background(),
		`SELECT id, status, attempts, next_attempt_at, last_error FROM outbox.outbox_events WHERE event_type = $1 ORDER BY id`, eventType)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []event
	for rows.Next() {
		var e event
		if err := rows.Scan(&e.ID, &e.Status, &e.Attempts, &e.Next, &e.LastErr); err != nil {
			t.Fatal(err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// publish записывает события в одной транзакции.
func publish(t *testing.T, db *pgxpool.Pool, eventType string, payloads ...any) {
	t.Helper()
	pub := NewPublisher(db)
	err := postgres.NewTxManager(db).Do(context.Background(), func(ctx context.Context) error {
		for _, p := range payloads {
			if err := pub.Publish(ctx, eventType, p); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// drain несколько раз обрабатывает очередь: нужные события не обязательно попадут в первую пачку.
func drain(t *testing.T, p *Processor) {
	t.Helper()
	for range 20 {
		if err := p.ProcessDue(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPublishJoinsTransaction(t *testing.T) {
	db := testdb.Connect(t)
	ctx := context.Background()
	pub := NewPublisher(db)
	tm := postgres.NewTxManager(db)
	typ := testdb.Unique("test.tx")

	// Откат транзакции — события нет.
	boom := errors.New("ошибка сценария")
	err := tm.Do(ctx, func(ctx context.Context) error {
		if err := pub.Publish(ctx, typ, map[string]int{"n": 1}); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) || len(eventsOf(t, db, typ)) != 0 {
		t.Fatalf("после отката событие осталось: err=%v", err)
	}

	// Фиксация — событие появилось вместе с транзакцией, payload в JSON.
	if err := tm.Do(ctx, func(ctx context.Context) error {
		return pub.Publish(ctx, typ, map[string]int{"n": 2})
	}); err != nil {
		t.Fatal(err)
	}
	evs := eventsOf(t, db, typ)
	if len(evs) != 1 || evs[0].Status != "pending" || evs[0].Attempts != 0 {
		t.Fatalf("после фиксации: %+v", evs)
	}
	var n int
	if err := db.QueryRow(ctx, `SELECT (payload->>'n')::int FROM outbox.outbox_events WHERE id = $1`, evs[0].ID).Scan(&n); err != nil || n != 2 {
		t.Fatalf("payload: n=%d err=%v", n, err)
	}

	// Вне транзакции — ошибка, ничего не записано.
	if err := pub.Publish(ctx, typ, 1); !errors.Is(err, ErrNoTransaction) {
		t.Fatalf("вне транзакции ожидалась ErrNoTransaction, получили %v", err)
	}
	if len(eventsOf(t, db, typ)) != 1 {
		t.Fatal("вне транзакции событие записалось")
	}

	// Payload, который нельзя закодировать в JSON, — ошибка.
	err = tm.Do(ctx, func(ctx context.Context) error { return pub.Publish(ctx, typ, make(chan int)) })
	if err == nil {
		t.Fatal("ожидалась ошибка кодирования payload")
	}
}

func TestProcessorDeliversInIDOrder(t *testing.T) {
	db := testdb.Connect(t)
	typA, typB := testdb.Unique("test.order.a"), testdb.Unique("test.order.b")
	// Чередуем два типа в одной транзакции.
	pub := NewPublisher(db)
	err := postgres.NewTxManager(db).Do(context.Background(), func(ctx context.Context) error {
		for i := range 6 {
			typ := typA
			if i%2 == 1 {
				typ = typB
			}
			if err := pub.Publish(ctx, typ, map[string]int{"n": i}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	var got []int
	h := func(_ context.Context, payload []byte) error {
		var p struct{ N int }
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		got = append(got, p.N)
		return nil
	}
	proc := NewProcessor(db)
	proc.Subscribe(typA, h)
	proc.Subscribe(typB, h)
	drain(t, proc)

	if len(got) != 6 {
		t.Fatalf("доставлено %v", got)
	}
	for i, n := range got {
		if n != i {
			t.Fatalf("нарушен порядок доставки: %v", got)
		}
	}
	for _, e := range append(eventsOf(t, db, typA), eventsOf(t, db, typB)...) {
		if e.Status != "done" || e.Attempts != 1 {
			t.Fatalf("событие не отмечено: %+v", e)
		}
	}
}

func TestProcessorRetriesWithBackoffAndGivesUp(t *testing.T) {
	db := testdb.Connect(t)
	ctx := context.Background()
	typ := testdb.Unique("test.retry")
	publish(t, db, typ, map[string]int{"n": 1})

	calls := 0
	proc := NewProcessor(db)
	proc.Subscribe(typ, func(context.Context, []byte) error {
		calls++
		return errors.New("SMTP недоступен")
	})

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		before := time.Now()
		drain(t, proc)
		e := eventsOf(t, db, typ)[0]
		if e.Attempts != attempt || calls != attempt || e.LastErr != "SMTP недоступен" {
			t.Fatalf("попытка %d: %+v, вызовов %d", attempt, e, calls)
		}
		if attempt < maxAttempts {
			// Повтор через attempt² минут.
			want := time.Duration(attempt*attempt) * time.Minute
			if delay := e.Next.Sub(before); e.Status != "pending" || delay < want-5*time.Second || delay > want+5*time.Second {
				t.Fatalf("попытка %d: status=%s, повтор через %v, ждали %v", attempt, e.Status, delay, want)
			}
			// Подгоняем срок повтора, чтобы не ждать минуты.
			if _, err := db.Exec(ctx, `UPDATE outbox.outbox_events SET next_attempt_at = now() WHERE id = $1`, e.ID); err != nil {
				t.Fatal(err)
			}
		} else if e.Status != "failed" {
			t.Fatalf("после %d попыток status=%s, ждали failed", attempt, e.Status)
		}
	}

	// Брошенное событие больше не берётся.
	drain(t, proc)
	if calls != maxAttempts {
		t.Fatalf("событие failed обработано ещё раз: вызовов %d", calls)
	}
}

func TestProcessorRecoversFromPanicAndSucceedsOnRetry(t *testing.T) {
	db := testdb.Connect(t)
	ctx := context.Background()
	typ := testdb.Unique("test.panic")
	publish(t, db, typ, 1)

	calls := 0
	proc := NewProcessor(db)
	proc.Subscribe(typ, func(context.Context, []byte) error {
		calls++
		if calls == 1 {
			panic("сбой")
		}
		return nil
	})
	drain(t, proc)
	if e := eventsOf(t, db, typ)[0]; e.Status != "pending" || e.LastErr == "" {
		t.Fatalf("после паники: %+v", e)
	}
	if _, err := db.Exec(ctx, `UPDATE outbox.outbox_events SET next_attempt_at = now() WHERE event_type = $1`, typ); err != nil {
		t.Fatal(err)
	}
	drain(t, proc)
	if e := eventsOf(t, db, typ)[0]; e.Status != "done" || e.Attempts != 2 || e.LastErr != "" {
		t.Fatalf("после успешного повтора: %+v", e)
	}
}

func TestProcessorMarksUnsubscribedDone(t *testing.T) {
	db := testdb.Connect(t)
	typ := testdb.Unique("test.nobody")
	publish(t, db, typ, 1, 2)
	drain(t, NewProcessor(db))
	evs := eventsOf(t, db, typ)
	if len(evs) != 2 || evs[0].Status != "done" || evs[1].Status != "done" {
		t.Fatalf("события без подписчиков: %+v", evs)
	}
}

// Два обработчика работают одновременно (как два экземпляра приложения):
// SKIP LOCKED и аренда не дают доставить одно событие дважды.
func TestConcurrentProcessorsDeliverOnce(t *testing.T) {
	db := testdb.Connect(t)
	typ := testdb.Unique("test.concurrent")
	const total = 60
	payloads := make([]any, total)
	for i := range payloads {
		payloads[i] = map[string]int{"n": i}
	}
	publish(t, db, typ, payloads...)

	var mu sync.Mutex
	delivered := map[int]int{}
	h := func(_ context.Context, payload []byte) error {
		var p struct{ N int }
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		time.Sleep(2 * time.Millisecond) // чтобы проходы обработчиков пересекались
		mu.Lock()
		delivered[p.N]++
		mu.Unlock()
		return nil
	}

	var wg sync.WaitGroup
	errc := make(chan error, 2)
	for range 2 {
		proc := NewProcessor(db)
		proc.Subscribe(typ, h)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 30 {
				if err := proc.ProcessDue(context.Background()); err != nil {
					errc <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errc)
	for err := range errc {
		t.Fatal(err)
	}

	if len(delivered) != total {
		t.Fatalf("доставлено %d событий из %d", len(delivered), total)
	}
	for n, c := range delivered {
		if c != 1 {
			t.Fatalf("событие %d доставлено %d раз", n, c)
		}
	}
	for _, e := range eventsOf(t, db, typ) {
		if e.Status != "done" || e.Attempts != 1 {
			t.Fatalf("событие взято повторно: %+v", e)
		}
	}
}

func TestSubscribeTwicePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("повторная подписка должна паниковать")
		}
	}()
	p := &Processor{handlers: map[string]Handler{}}
	h := func(context.Context, []byte) error { return nil }
	p.Subscribe("x", h)
	p.Subscribe("x", h)
}
