-- Outbox событий (infrastructure/outbox). Модуль технический: запросы лежат здесь,
-- чтобы весь SQL проекта был в queries/, а код генерировал sqlc.

-- name: InsertOutboxEvent :exec
INSERT INTO outbox.outbox_events (event_type, payload) VALUES ($1, $2);

-- Забираем события, которым пора обрабатываться, и сразу откладываем их на время аренды:
-- если обработчик упадёт посреди работы, событие вернётся в очередь само.
-- SKIP LOCKED — несколько обработчиков (или экземпляров приложения) не возьмут одно событие дважды.
-- name: ClaimDueOutboxEvents :many
UPDATE outbox.outbox_events
SET attempts = attempts + 1,
    next_attempt_at = now() + make_interval(secs => sqlc.arg(lease_seconds)::int)
WHERE id IN (
    SELECT id FROM outbox.outbox_events
    WHERE status = 'pending' AND next_attempt_at <= now()
    ORDER BY id
    LIMIT sqlc.arg(batch_size)
    FOR UPDATE SKIP LOCKED
)
RETURNING id, event_type, payload, attempts;

-- name: MarkOutboxEventDone :exec
UPDATE outbox.outbox_events SET status = 'done', processed_at = now(), last_error = '' WHERE id = $1;

-- Неудачная попытка: либо повтор позже (status = pending), либо отказ (status = failed).
-- name: MarkOutboxEventFailed :exec
UPDATE outbox.outbox_events
SET status = sqlc.arg(status),
    last_error = sqlc.arg(last_error),
    next_attempt_at = sqlc.arg(next_attempt_at),
    processed_at = CASE WHEN sqlc.arg(status)::text = 'failed' THEN now() END
WHERE id = sqlc.arg(id);
