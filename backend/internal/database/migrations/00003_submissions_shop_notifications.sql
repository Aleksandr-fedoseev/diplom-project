-- +goose Up
-- Заявки и модерация, магазин и заказы, очередь email-уведомлений.

-- Заявка пользователя: новый объект или дополнение (фото, сведения) к существующему.
CREATE TABLE submissions (
    id           BIGSERIAL PRIMARY KEY,
    author_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind         TEXT   NOT NULL CHECK (kind IN ('new_object', 'addition')),
    -- Для дополнения — объект, к которому оно относится; после одобрения новой заявки — созданный объект.
    object_id    BIGINT REFERENCES heritage_objects (id) ON DELETE SET NULL,
    title        TEXT   NOT NULL,
    description  TEXT   NOT NULL DEFAULT '',
    address      TEXT   NOT NULL DEFAULT '',
    location     GEOGRAPHY(POINT, 4326),
    period       TEXT   NOT NULL DEFAULT '', -- период или дата, к которым относятся сведения
    source       TEXT   NOT NULL,            -- источник сведений: архив, книга, личный архив и т. п.
    status       TEXT   NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    -- Что у дополнения есть объект, проверяет модуль заявок: CHECK здесь мешал бы удалить
    -- объект — ON DELETE SET NULL оставил бы дополнение без объекта и нарушил бы проверку.
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX submissions_author_idx ON submissions (author_id);
CREATE INDEX submissions_pending_idx ON submissions (created_at) WHERE status = 'pending';

-- Решение модератора по заявке: одно на заявку.
CREATE TABLE moderation_decisions (
    id            BIGSERIAL PRIMARY KEY,
    submission_id BIGINT NOT NULL UNIQUE REFERENCES submissions (id) ON DELETE CASCADE,
    moderator_id  BIGINT REFERENCES users (id) ON DELETE SET NULL,
    decision      TEXT   NOT NULL CHECK (decision IN ('approved', 'rejected')),
    reason        TEXT   NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (decision = 'approved' OR reason <> '')
);

-- Медиафайл прикреплён к объекту или к заявке. После одобрения заявки файлы переходят к объекту.
ALTER TABLE media_files
    ADD COLUMN submission_id BIGINT REFERENCES submissions (id) ON DELETE CASCADE,
    ADD COLUMN year_taken    TEXT NOT NULL DEFAULT '',
    ADD COLUMN source        TEXT NOT NULL DEFAULT '',
    ADD CONSTRAINT media_files_owner_check CHECK (object_id IS NOT NULL OR submission_id IS NOT NULL);

CREATE INDEX media_files_submission_idx ON media_files (submission_id);

-- Цены храним в копейках целым числом: у дробных чисел бывают ошибки округления.
CREATE TABLE products (
    id           BIGSERIAL PRIMARY KEY,
    slug         TEXT    NOT NULL UNIQUE,
    title        TEXT    NOT NULL,
    description  TEXT    NOT NULL DEFAULT '',
    category     TEXT    NOT NULL DEFAULT '',
    price_kopecks BIGINT NOT NULL CHECK (price_kopecks > 0),
    stock        INT     NOT NULL DEFAULT 0 CHECK (stock >= 0),
    image_url    TEXT    NOT NULL DEFAULT '',
    is_active    BOOLEAN NOT NULL DEFAULT TRUE, -- снятый с продажи товар остаётся в старых заказах
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE cart_items (
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    quantity   INT    NOT NULL CHECK (quantity > 0),
    added_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, product_id)
);

-- Статусы заказа: ждёт оплаты → оплачен → отправлен → выполнен; отменён — из «ждёт оплаты» или «оплачен».
CREATE TABLE orders (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT REFERENCES users (id) ON DELETE SET NULL,
    total_kopecks  BIGINT NOT NULL CHECK (total_kopecks > 0),
    status         TEXT   NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'shipped', 'completed', 'cancelled')),
    payment_id     TEXT UNIQUE, -- id платежа в ЮKassa
    payment_url    TEXT NOT NULL DEFAULT '', -- страница оплаты, куда отправляем покупателя
    delivery_address TEXT NOT NULL DEFAULT '',
    phone          TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX orders_user_idx ON orders (user_id);

CREATE TABLE order_items (
    order_id      BIGINT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id    BIGINT NOT NULL REFERENCES products (id),
    title         TEXT   NOT NULL, -- название на момент заказа
    quantity      INT    NOT NULL CHECK (quantity > 0),
    price_kopecks BIGINT NOT NULL, -- цена на момент заказа
    PRIMARY KEY (order_id, product_id)
);

-- Очередь писем: модуль кладёт сюда задачу, фоновый обработчик отправляет и при ошибке повторяет.
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

-- Стартовые товары из прототипа фронтенда (src/data/shop.ts).
INSERT INTO products (slug, title, description, category, price_kopecks, stock, image_url) VALUES
    ('merch-hoodie-astrakhan', 'Худи «Астрахань архитектурная»',
     'Черный плотный футер с авторской графикой фасадов купеческой Астрахани. 100% хлопок.',
     'clothes', 450000, 20, '/merch/hoodie.jpg'),
    ('merch-tshirt-volga', 'Футболка «Волга и Каспий»',
     'Минималистичный принт фарватера Волги и Каспия по старинным архивным картам. Плотный хлопок.',
     'clothes', 250000, 30, '/merch/tshirt.jpg'),
    ('merch-shopper-heart', 'Сумка-шоппер «Астрахань в сердце»',
     'Прочный холщовый шоппер из парусины с графикой исторической стрелки Волги и Кутума.',
     'accessories', 120000, 40, '/merch/shopper.jpg'),
    ('merch-cap-minimal', 'Кепка с вышивкой «Астрахань»',
     'Черная хлопковая бейсболка с лаконичной белой вышивкой «Астрахань» и регулируемым фиксатором.',
     'hats', 180000, 25, '/merch/cap.jpg'),
    ('merch-notebook-map', 'Блокнот «Карта региона»',
     'Твердый тактильный переплет, фактурная крафтовая бумага и тиснение карты Астраханского края.',
     'accessories', 90000, 50, '/merch/notebook.jpg'),
    ('merch-book-heritage', 'Книга «Архитектурное наследие Астрахани»',
     'Научное издание с обмерными чертежами, архивными фото ГАОО и паспортами памятников архитектуры.',
     'books', 195000, 15, '/heritage/dom-gubina/cover.jpg');

-- +goose Down
DROP TABLE email_jobs;
DROP TABLE order_items;
DROP TABLE orders;
DROP TABLE cart_items;
DROP TABLE products;
ALTER TABLE media_files
    DROP CONSTRAINT media_files_owner_check,
    DROP COLUMN source,
    DROP COLUMN year_taken,
    DROP COLUMN submission_id;
DROP TABLE moderation_decisions;
DROP TABLE submissions;
