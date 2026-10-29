-- Запросы модуля маршрутов (repository/postgres/routes.go).
-- Линия маршрута приходит от маршрутизатора в GeoJSON и хранится как geography.

-- name: CreateRoute :one
INSERT INTO routes (slug, title, description, owner_id, distance_m, duration_s, geometry)
VALUES (sqlc.arg(slug), sqlc.arg(title), sqlc.arg(description), sqlc.narg(owner_id),
        sqlc.arg(distance_m), sqlc.arg(duration_s), ST_GeomFromGeoJSON(sqlc.arg(geometry)::text)::geography)
RETURNING id;

-- name: CreateRoutePoint :exec
INSERT INTO route_points (route_id, position, object_id) VALUES ($1, $2, $3);

-- Списки без линий: для списка линии не нужны, они тяжёлые.
-- name: ListThematicRoutes :many
SELECT id, slug, title, description, owner_id, distance_m, duration_s, created_at
FROM routes WHERE owner_id IS NULL ORDER BY title;

-- name: ListRoutesByOwner :many
SELECT id, slug, title, description, owner_id, distance_m, duration_s, created_at
FROM routes WHERE owner_id = $1 ORDER BY created_at DESC;

-- name: GetRouteBySlug :one
SELECT id, slug, title, description, owner_id, distance_m, duration_s, created_at,
       ST_AsGeoJSON(geometry)::text AS geometry
FROM routes WHERE slug = $1;

-- Сами объекты остановок модуль получает у каталога по object_id.
-- Порядок столбцов не как в таблице: так sqlc создаёт отдельную структуру строки
-- и не добавляет модель RoutePoint в общий models.go.
-- name: ListRoutePoints :many
SELECT route_id, object_id, position FROM route_points
WHERE route_id = ANY(sqlc.arg(route_ids)::bigint[])
ORDER BY route_id, position;

-- name: GetRouteOwner :one
SELECT owner_id FROM routes WHERE id = $1;

-- name: DeleteRoute :exec
DELETE FROM routes WHERE id = $1;
