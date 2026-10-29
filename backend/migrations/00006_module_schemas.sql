-- +goose Up
-- У каждого модуля своя схема: границы модулей видны прямо в базе
-- (auth.users, catalog.heritage_objects …). Внешние ключи между схемами продолжают работать,
-- а модули по-прежнему обращаются к чужим данным только через сервис владельца.
-- Последовательности и индексы переезжают вместе с таблицами.

CREATE SCHEMA auth;
ALTER TABLE roles SET SCHEMA auth;
ALTER TABLE users SET SCHEMA auth;

CREATE SCHEMA catalog;
ALTER TABLE categories SET SCHEMA catalog;
ALTER TABLE heritage_objects SET SCHEMA catalog;

CREATE SCHEMA media;
ALTER TABLE media_files SET SCHEMA media;

CREATE SCHEMA route;
ALTER TABLE routes SET SCHEMA route;
ALTER TABLE route_points SET SCHEMA route;

CREATE SCHEMA submission;
ALTER TABLE submissions SET SCHEMA submission;
ALTER TABLE moderation_decisions SET SCHEMA submission;

CREATE SCHEMA shop;
ALTER TABLE products SET SCHEMA shop;
ALTER TABLE cart_items SET SCHEMA shop;
ALTER TABLE orders SET SCHEMA shop;
ALTER TABLE order_items SET SCHEMA shop;

CREATE SCHEMA outbox;
ALTER TABLE outbox_events SET SCHEMA outbox;

-- Очередь готовых писем заменена событиями outbox: письма теперь собирает модуль notification.
DROP TABLE email_jobs;

-- +goose Down
CREATE TABLE email_jobs (
    id              BIGSERIAL PRIMARY KEY,
    to_email        TEXT NOT NULL,
    subject         TEXT NOT NULL,
    body            TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    attempts        INT  NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error      TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at         TIMESTAMPTZ
);
CREATE INDEX email_jobs_due_idx ON email_jobs (next_attempt_at) WHERE status = 'pending';

ALTER TABLE outbox.outbox_events SET SCHEMA public;
ALTER TABLE shop.order_items SET SCHEMA public;
ALTER TABLE shop.orders SET SCHEMA public;
ALTER TABLE shop.cart_items SET SCHEMA public;
ALTER TABLE shop.products SET SCHEMA public;
ALTER TABLE submission.moderation_decisions SET SCHEMA public;
ALTER TABLE submission.submissions SET SCHEMA public;
ALTER TABLE route.route_points SET SCHEMA public;
ALTER TABLE route.routes SET SCHEMA public;
ALTER TABLE media.media_files SET SCHEMA public;
ALTER TABLE catalog.heritage_objects SET SCHEMA public;
ALTER TABLE catalog.categories SET SCHEMA public;
ALTER TABLE auth.users SET SCHEMA public;
ALTER TABLE auth.roles SET SCHEMA public;
DROP SCHEMA outbox, shop, submission, route, media, catalog, auth;
