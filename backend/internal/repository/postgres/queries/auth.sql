-- Пользователи (модуль auth).

-- name: CreateUser :one
INSERT INTO auth.users (email, password_hash, name, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM auth.users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM auth.users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM auth.users ORDER BY id LIMIT $1 OFFSET $2;

-- name: SetUserRole :one
UPDATE auth.users SET role = $2 WHERE id = $1 RETURNING *;

-- name: SetUserBlocked :one
UPDATE auth.users SET is_blocked = $2 WHERE id = $1 RETURNING *;
