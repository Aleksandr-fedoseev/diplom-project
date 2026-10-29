-- name: EnqueueEmail :exec
INSERT INTO email_jobs (to_email, subject, body) VALUES ($1, $2, $3);

-- Забираем письма, которым пора уйти, и сразу откладываем их на 5 минут (аренда):
-- если обработчик упадёт посреди отправки, письмо вернётся в очередь само.
-- SKIP LOCKED — несколько экземпляров приложения не возьмут одно письмо дважды.
-- name: ClaimDueEmails :many
UPDATE email_jobs
SET attempts = attempts + 1, next_attempt_at = now() + interval '5 minutes'
WHERE id IN (
    SELECT id FROM email_jobs
    WHERE status = 'pending' AND next_attempt_at <= now()
    ORDER BY next_attempt_at
    LIMIT sqlc.arg(batch_size)
    FOR UPDATE SKIP LOCKED
)
RETURNING id, to_email, subject, body, attempts;

-- name: MarkEmailSent :exec
UPDATE email_jobs SET status = 'sent', sent_at = now(), last_error = '' WHERE id = $1;

-- name: MarkEmailFailed :exec
UPDATE email_jobs
SET status = sqlc.arg(status), last_error = sqlc.arg(last_error), next_attempt_at = sqlc.arg(next_attempt_at)
WHERE id = sqlc.arg(id);
