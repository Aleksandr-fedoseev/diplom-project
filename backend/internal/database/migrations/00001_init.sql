-- +goose Up
-- Первая миграция: пользователи, каталог объектов, медиафайлы и маршруты.
-- Заявки, магазин и заказы добавятся следующими миграциями.

CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE roles (
    code  TEXT PRIMARY KEY,          -- user | moderator | admin
    name  TEXT NOT NULL,
    level INT  NOT NULL UNIQUE       -- чем больше, тем больше прав
);

INSERT INTO roles (code, name, level) VALUES
    ('user',      'Пользователь',  1),
    ('moderator', 'Модератор',     2),
    ('admin',     'Администратор', 3);

CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    name          TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'user' REFERENCES roles (code),
    is_blocked    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE categories (
    id   SERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);

CREATE TABLE heritage_objects (
    id                  BIGSERIAL PRIMARY KEY,
    slug                TEXT    NOT NULL UNIQUE,
    title               TEXT    NOT NULL,
    category_id         INT     NOT NULL REFERENCES categories (id),
    protection_status   TEXT    NOT NULL
        CHECK (protection_status IN ('federal', 'regional', 'identified', 'lost')),
    registry_number     TEXT,
    year_built          TEXT,
    architect           TEXT,
    architectural_style TEXT,
    address             TEXT    NOT NULL,
    -- geography хранит точку в градусах WGS 84 и считает расстояния в метрах.
    location            GEOGRAPHY(POINT, 4326) NOT NULL,
    short_description   TEXT    NOT NULL DEFAULT '',
    full_history        TEXT    NOT NULL DEFAULT '',
    tags                TEXT[]  NOT NULL DEFAULT '{}',
    is_published        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Пространственный индекс: ускоряет поиск «объекты рядом» и «объекты в видимой области карты».
CREATE INDEX heritage_objects_location_idx ON heritage_objects USING GIST (location);
CREATE INDEX heritage_objects_category_idx ON heritage_objects (category_id);

-- Сами файлы лежат в MinIO, здесь только сведения о них и ключ в хранилище.
CREATE TABLE media_files (
    id            BIGSERIAL PRIMARY KEY,
    object_id     BIGINT REFERENCES heritage_objects (id) ON DELETE CASCADE,
    kind          TEXT   NOT NULL CHECK (kind IN ('photo', 'archive_photo', 'model3d', 'audio')),
    storage_key   TEXT   NOT NULL UNIQUE,
    original_name TEXT   NOT NULL,
    content_type  TEXT   NOT NULL,
    size_bytes    BIGINT NOT NULL,
    caption       TEXT   NOT NULL DEFAULT '',
    uploaded_by   BIGINT REFERENCES users (id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX media_files_object_idx ON media_files (object_id);

CREATE TABLE routes (
    id           BIGSERIAL PRIMARY KEY,
    slug         TEXT NOT NULL UNIQUE,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    -- NULL — тематический маршрут администратора, иначе — сохранённый маршрут пользователя.
    owner_id     BIGINT REFERENCES users (id) ON DELETE CASCADE,
    distance_m   DOUBLE PRECISION NOT NULL,
    duration_s   DOUBLE PRECISION NOT NULL,
    -- Линия, построенная OSRM по улицам; хранится, чтобы не пересчитывать при каждом показе.
    geometry     GEOGRAPHY(LINESTRING, 4326) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE route_points (
    route_id  BIGINT NOT NULL REFERENCES routes (id) ON DELETE CASCADE,
    position  INT    NOT NULL,
    object_id BIGINT NOT NULL REFERENCES heritage_objects (id),
    PRIMARY KEY (route_id, position)
);

-- +goose Down
DROP TABLE route_points;
DROP TABLE routes;
DROP TABLE media_files;
DROP TABLE heritage_objects;
DROP TABLE categories;
DROP TABLE users;
DROP TABLE roles;
