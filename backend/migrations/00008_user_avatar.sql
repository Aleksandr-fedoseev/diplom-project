-- +goose Up
-- Аватар пользователя: ключ файла в S3-хранилище; NULL — аватара нет.
ALTER TABLE auth.users ADD COLUMN avatar_key TEXT;

-- +goose Down
ALTER TABLE auth.users DROP COLUMN avatar_key;
