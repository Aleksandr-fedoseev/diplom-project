-- Запросы модуля медиафайлов. Сами файлы — в S3-хранилище, здесь только сведения о них.

-- name: CreateMediaFile :one
INSERT INTO media_files
    (object_id, submission_id, kind, storage_key, original_name, content_type, size_bytes,
     caption, year_taken, source_note, source, projection, h_angle, author, license, uploaded_by)
VALUES (
    sqlc.narg(object_id), sqlc.narg(submission_id), sqlc.arg(kind), sqlc.arg(storage_key),
    sqlc.arg(original_name), sqlc.arg(content_type), sqlc.arg(size_bytes),
    sqlc.arg(caption), sqlc.arg(year_taken), sqlc.arg(source_note), sqlc.arg(source),
    sqlc.narg(projection), sqlc.narg(h_angle), sqlc.arg(author), sqlc.arg(license), sqlc.narg(uploaded_by)
)
RETURNING *;

-- name: ListMediaFilesByObject :many
SELECT * FROM media_files WHERE object_id = $1 ORDER BY kind, id;

-- name: ListMediaFilesBySubmission :many
SELECT * FROM media_files WHERE submission_id = $1 ORDER BY id;

-- name: GetMediaFile :one
SELECT * FROM media_files WHERE id = $1;

-- Одобренная заявка передаёт свои файлы объекту; связь с заявкой остаётся для истории.
-- name: AttachMediaFilesToObject :exec
UPDATE media_files SET object_id = sqlc.arg(object_id) WHERE submission_id = sqlc.arg(submission_id);

-- name: DeleteMediaFile :execrows
DELETE FROM media_files WHERE id = $1;
