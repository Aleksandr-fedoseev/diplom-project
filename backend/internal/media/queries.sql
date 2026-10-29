-- name: CreateMediaFile :one
INSERT INTO media_files
    (object_id, submission_id, kind, storage_key, original_name, content_type, size_bytes,
     caption, year_taken, source_note, uploaded_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: ListMediaFilesByObject :many
SELECT * FROM media_files WHERE object_id = $1 ORDER BY kind, id;

-- name: ListMediaFilesBySubmission :many
SELECT * FROM media_files WHERE submission_id = $1 ORDER BY id;

-- name: GetMediaFile :one
SELECT * FROM media_files WHERE id = $1;

-- Одобренная заявка передаёт свои файлы объекту; связь с заявкой остаётся для истории.
-- name: AttachSubmissionMedia :exec
UPDATE media_files SET object_id = sqlc.arg(object_id) WHERE submission_id = sqlc.arg(submission_id);

-- name: DeleteMediaFile :exec
DELETE FROM media_files WHERE id = $1;
