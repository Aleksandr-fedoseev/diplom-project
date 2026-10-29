-- Заявки и решения модераторов (модуль submission).
-- Модуль владеет таблицами submissions и moderation_decisions; объекты, файлы и
-- пользователи — через другие модули.

-- Точка собирается из долготы и широты; если координат нет, ST_MakePoint(NULL, NULL) даёт NULL.
-- name: CreateSubmission :one
INSERT INTO submissions (author_id, kind, object_id, title, description, address, location, period, source)
VALUES (
    sqlc.arg(author_id), sqlc.arg(kind), sqlc.narg(object_id), sqlc.arg(title), sqlc.arg(description),
    sqlc.arg(address),
    ST_SetSRID(ST_MakePoint(sqlc.narg(lng)::float8, sqlc.narg(lat)::float8), 4326)::geography,
    sqlc.arg(period), sqlc.arg(source)
)
RETURNING id;

-- Заявка вместе с решением модератора (если оно уже есть).
-- Координаты необязательны: has_location говорит, заданы ли они, lat/lng без точки — нули.
-- name: GetSubmission :one
SELECT
    s.id, s.author_id, s.kind, s.object_id, s.title, s.description, s.address,
    (s.location IS NOT NULL)::bool AS has_location,
    COALESCE(ST_Y(s.location::geometry), 0)::float8 AS lat,
    COALESCE(ST_X(s.location::geometry), 0)::float8 AS lng,
    s.period, s.source, s.status, s.created_at,
    d.decision, d.reason, d.moderator_id, d.created_at AS decided_at
FROM submissions s
LEFT JOIN moderation_decisions d ON d.submission_id = s.id
WHERE s.id = $1;

-- Свои заявки пользователя, новые первыми.
-- name: ListSubmissionsByAuthor :many
SELECT
    s.id, s.author_id, s.kind, s.object_id, s.title, s.description, s.address,
    (s.location IS NOT NULL)::bool AS has_location,
    COALESCE(ST_Y(s.location::geometry), 0)::float8 AS lat,
    COALESCE(ST_X(s.location::geometry), 0)::float8 AS lng,
    s.period, s.source, s.status, s.created_at,
    d.decision, d.reason, d.moderator_id, d.created_at AS decided_at
FROM submissions s
LEFT JOIN moderation_decisions d ON d.submission_id = s.id
WHERE s.author_id = $1
ORDER BY s.created_at DESC, s.id DESC;

-- Очередь модерации: старые заявки первыми, чтобы никто не ждал дольше других.
-- Для ожидающих заявок работает частичный индекс submissions_pending_idx.
-- name: ListSubmissionsByStatus :many
SELECT
    s.id, s.author_id, s.kind, s.object_id, s.title, s.description, s.address,
    (s.location IS NOT NULL)::bool AS has_location,
    COALESCE(ST_Y(s.location::geometry), 0)::float8 AS lat,
    COALESCE(ST_X(s.location::geometry), 0)::float8 AS lng,
    s.period, s.source, s.status, s.created_at,
    d.decision, d.reason, d.moderator_id, d.created_at AS decided_at
FROM submissions s
LEFT JOIN moderation_decisions d ON d.submission_id = s.id
WHERE s.status = $1
ORDER BY s.created_at, s.id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- Блокирует строку заявки до конца транзакции: второй модератор, решающий ту же заявку,
-- дождётся первого и увидит, что она уже не ожидает решения.
-- Решение не читаем: у ожидающей заявки его нет, а после решения заявку уже не меняют.
-- name: LockSubmission :one
SELECT
    id, author_id, kind, object_id, title, description, address,
    (location IS NOT NULL)::bool AS has_location,
    COALESCE(ST_Y(location::geometry), 0)::float8 AS lat,
    COALESCE(ST_X(location::geometry), 0)::float8 AS lng,
    period, source, status, created_at
FROM submissions
WHERE id = $1
FOR UPDATE;

-- object_id меняется только при одобрении (созданный или существующий объект), иначе остаётся прежним.
-- name: SetSubmissionStatus :execrows
UPDATE submissions
SET status = sqlc.arg(status), object_id = COALESCE(sqlc.narg(object_id), object_id)
WHERE id = sqlc.arg(id);

-- name: CreateModerationDecision :exec
INSERT INTO moderation_decisions (submission_id, moderator_id, decision, reason)
VALUES ($1, $2, $3, $4);
