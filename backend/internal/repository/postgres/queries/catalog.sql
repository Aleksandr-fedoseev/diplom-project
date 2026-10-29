-- Запросы модуля каталога: категории и объекты наследия (PostGIS для поиска на карте).

-- name: ListCategories :many
SELECT id, slug, name FROM categories ORDER BY id;

-- name: CreateCategory :one
INSERT INTO categories (slug, name) VALUES ($1, $2) RETURNING id, slug, name;

-- name: UpdateCategory :one
UPDATE categories SET slug = $2, name = $3 WHERE id = $1 RETURNING id, slug, name;

-- name: DeleteCategory :execrows
DELETE FROM categories WHERE id = $1;

-- Список объектов. Фильтры необязательны: фильтр со значением NULL не применяется.
-- Координаты достаём из geography: ST_Y — широта, ST_X — долгота.
-- bbox — видимая область карты: && сравнивает с прямоугольником по GIST-индексу.
-- name: ListObjects :many
SELECT
    o.id, o.slug, o.title, c.slug AS category, o.protection_status, o.registry_number, o.year_built,
    o.architect, o.architectural_style, o.address,
    ST_Y(o.location::geometry)::float8 AS lat,
    ST_X(o.location::geometry)::float8 AS lng,
    o.short_description, o.tags, o.is_published, o.updated_at
FROM heritage_objects o
JOIN categories c ON c.id = o.category_id
WHERE (sqlc.arg(include_unpublished)::bool OR o.is_published)
  AND (sqlc.narg(category)::text IS NULL OR c.slug = sqlc.narg(category)::text)
  AND (sqlc.narg(status)::text IS NULL OR o.protection_status = sqlc.narg(status)::text)
  AND (sqlc.narg(query)::text IS NULL
       OR o.title ILIKE '%' || sqlc.narg(query)::text || '%'
       OR o.address ILIKE '%' || sqlc.narg(query)::text || '%')
  AND (sqlc.narg(min_lng)::float8 IS NULL
       OR o.location && ST_MakeEnvelope(sqlc.narg(min_lng)::float8, sqlc.narg(min_lat)::float8,
                                        sqlc.narg(max_lng)::float8, sqlc.narg(max_lat)::float8, 4326)::geography)
ORDER BY o.title
LIMIT 500;

-- То же, но только объекты в радиусе от точки, ближайшие первыми.
-- ST_DWithin использует GIST-индекс, поэтому не перебирает все объекты.
-- name: ListObjectsNear :many
SELECT
    o.id, o.slug, o.title, c.slug AS category, o.protection_status, o.registry_number, o.year_built,
    o.architect, o.architectural_style, o.address,
    ST_Y(o.location::geometry)::float8 AS lat,
    ST_X(o.location::geometry)::float8 AS lng,
    o.short_description, o.tags, o.is_published, o.updated_at,
    ST_Distance(o.location, ST_SetSRID(ST_MakePoint(sqlc.arg(lng)::float8, sqlc.arg(lat)::float8), 4326)::geography)::float8 AS distance_m
FROM heritage_objects o
JOIN categories c ON c.id = o.category_id
WHERE (sqlc.arg(include_unpublished)::bool OR o.is_published)
  AND (sqlc.narg(category)::text IS NULL OR c.slug = sqlc.narg(category)::text)
  AND (sqlc.narg(status)::text IS NULL OR o.protection_status = sqlc.narg(status)::text)
  AND (sqlc.narg(query)::text IS NULL
       OR o.title ILIKE '%' || sqlc.narg(query)::text || '%'
       OR o.address ILIKE '%' || sqlc.narg(query)::text || '%')
  AND (sqlc.narg(min_lng)::float8 IS NULL
       OR o.location && ST_MakeEnvelope(sqlc.narg(min_lng)::float8, sqlc.narg(min_lat)::float8,
                                        sqlc.narg(max_lng)::float8, sqlc.narg(max_lat)::float8, 4326)::geography)
  AND ST_DWithin(o.location,
                 ST_SetSRID(ST_MakePoint(sqlc.arg(lng)::float8, sqlc.arg(lat)::float8), 4326)::geography,
                 sqlc.arg(radius_m)::float8)
ORDER BY distance_m
LIMIT 500;

-- name: GetObjectBySlug :one
SELECT
    o.id, o.slug, o.title, c.slug AS category, o.protection_status, o.registry_number, o.year_built,
    o.architect, o.architectural_style, o.address,
    ST_Y(o.location::geometry)::float8 AS lat,
    ST_X(o.location::geometry)::float8 AS lng,
    o.short_description, o.full_history, o.tags, o.is_published, o.updated_at
FROM heritage_objects o
JOIN categories c ON c.id = o.category_id
WHERE o.slug = $1;

-- name: CreateObject :exec
INSERT INTO heritage_objects
    (slug, title, category_id, protection_status, registry_number, year_built, architect,
     architectural_style, address, location, short_description, full_history, tags, is_published)
VALUES (
    sqlc.arg(slug), sqlc.arg(title), (SELECT cat.id FROM categories cat WHERE cat.slug = sqlc.arg(category)),
    sqlc.arg(protection_status), sqlc.narg(registry_number), sqlc.narg(year_built), sqlc.narg(architect),
    sqlc.narg(architectural_style), sqlc.arg(address),
    ST_SetSRID(ST_MakePoint(sqlc.arg(lng)::float8, sqlc.arg(lat)::float8), 4326)::geography,
    sqlc.arg(short_description), sqlc.arg(full_history), sqlc.arg(tags), sqlc.arg(is_published)
);

-- name: UpdateObject :execrows
UPDATE heritage_objects SET
    slug = sqlc.arg(slug),
    title = sqlc.arg(title),
    category_id = (SELECT cat.id FROM categories cat WHERE cat.slug = sqlc.arg(category)),
    protection_status = sqlc.arg(protection_status),
    registry_number = sqlc.narg(registry_number),
    year_built = sqlc.narg(year_built),
    architect = sqlc.narg(architect),
    architectural_style = sqlc.narg(architectural_style),
    address = sqlc.arg(address),
    location = ST_SetSRID(ST_MakePoint(sqlc.arg(lng)::float8, sqlc.arg(lat)::float8), 4326)::geography,
    short_description = sqlc.arg(short_description),
    full_history = sqlc.arg(full_history),
    tags = sqlc.arg(tags),
    is_published = sqlc.arg(is_published),
    updated_at = now()
WHERE heritage_objects.id = sqlc.arg(id);

-- name: DeleteObject :execrows
DELETE FROM heritage_objects WHERE id = $1;

-- Запросы для других модулей (маршруты, медиафайлы): они не читают таблицы каталога сами.

-- name: ObjectExists :one
SELECT EXISTS (SELECT 1 FROM heritage_objects WHERE id = $1);

-- name: ListPublishedPlacesBySlugs :many
SELECT id, slug, title,
       ST_Y(location::geometry)::float8 AS lat,
       ST_X(location::geometry)::float8 AS lng
FROM heritage_objects
WHERE slug = ANY(sqlc.arg(slugs)::text[]) AND is_published;

-- name: ListPlacesByIDs :many
SELECT id, slug, title,
       ST_Y(location::geometry)::float8 AS lat,
       ST_X(location::geometry)::float8 AS lng
FROM heritage_objects
WHERE id = ANY(sqlc.arg(ids)::bigint[]);
