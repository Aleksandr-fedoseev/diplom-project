-- +goose Up
-- Outbox: события модулей, записанные в той же транзакции, что и изменение данных.
-- Фоновый обработчик (infrastructure/outbox) доставляет их подписчикам, например модулю уведомлений.
CREATE TABLE outbox_events (
    id              BIGSERIAL PRIMARY KEY,
    event_type      TEXT        NOT NULL, -- имя события, например order.paid
    payload         JSONB       NOT NULL, -- данные события в JSON
    status          TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'done', 'failed')),
    attempts        INT         NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(), -- раньше этого времени событие не берём
    last_error      TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at    TIMESTAMPTZ -- когда событие доставлено (done) или брошено (failed)
);

-- Обработчик ищет только ожидающие события, срок которых наступил: частичный индекс
-- не растёт за счёт обработанных событий.
CREATE INDEX outbox_events_due_idx ON outbox_events (next_attempt_at, id) WHERE status = 'pending';

-- +goose Down
DROP TABLE outbox_events;
