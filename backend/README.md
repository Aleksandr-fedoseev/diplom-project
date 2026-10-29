# Бэкенд платформы наследия (Go, монолит)

Один Go-сервис по диаграмме развёртывания (`Diagrams/Deployment.png`). Данные — PostgreSQL 16 + PostGIS,
файлы — S3-хранилище (SeaweedFS, клиент minio-go), маршруты по улицам — OSRM, оплата — ЮKassa,
письма — SMTP.

## Архитектура

Модульный монолит, разложенный по слоям чистой архитектуры: одно приложение, один процесс, одна БД.
Код разделён по **слоям**, внутри каждого слоя — по **модулям**:

```
cmd/server/main.go          ← сигналы ОС, настройки, app.New(...).Run(ctx)
migrations/                 ← SQL-миграции goose (вшиты в бинарник) — они же схема для sqlc
internal/
├── app/                    ← сборка и жизненный цикл: единственное место, где создаются конкретные типы
│   ├── wiring.go           ← репозитории → сервисы → обработчики
│   └── adapters.go         ← переводчики между портами модулей
├── domain/<модуль>/        ← сущности, правила, ошибки, интерфейсы репозиториев и сервиса
├── service/<модуль>/       ← сценарии; техника и другие модули — через интерфейсы в ports.go
├── repository/postgres/    ← репозитории; SQL в queries/<модуль>.sql, код — sqlc в пакет sqlc
├── infrastructure/         ← техника без знания домена: postgres, s3, osrm, yookassa, smtp,
│                             security (bcrypt, JWT), outbox, config
└── transport/http/         ← роутер, middleware, обработчики модулей; ошибки домена → HTTP-коды
```

Направление зависимостей: `transport → service → domain ← repository/infrastructure`.
Домен импортирует только стандартную библиотеку. Модули не импортируют друг друга: каждый объявляет
маленький интерфейс в своём `ports.go`, а `app/adapters.go` подставляет реализацию другого модуля.

| Модуль | Что делает | Схема БД |
| --- | --- | --- |
| `auth` | регистрация, вход, JWT, роли, блокировка, управление пользователями | `auth` |
| `catalog` | объекты и категории, геопоиск PostGIS (радиус, видимая область карты) | `catalog` |
| `media` | фото, аудио, 3D и панорамы в S3; пометка «реконструкция ИИ» | `media` |
| `route` | тематические и свои маршруты, линия по улицам через OSRM | `route` |
| `submission` | заявки и модерация: одобрение с публикацией, отказ с причиной | `submission` |
| `shop` | товары, корзина, заказы, оплата ЮKassa | `shop` |
| `notification` | письма по событиям из outbox (без HTTP) | — |

**Транзакции.** Сервис открывает транзакцию через порт `TxManager.Do(ctx, fn)`; реализация
(`infrastructure/postgres`) кладёт её в `ctx`, а все репозитории берут соединение функцией `Conn`
(транзакция, если есть, иначе пул). Так одно одобрение заявки создаёт объект в каталоге, переносит
файлы и записывает решение атомарно, хотя это три разных модуля.

**События (outbox).** Побочные эффекты между модулями — через события, записанные в той же транзакции
(`outbox.outbox_events`). Фоновый обработчик доставляет их подписчикам (сейчас — модулю notification)
с повторами через 1, 4, 9, 16 минут; доставка «хотя бы один раз». События: `submission.approved`,
`submission.rejected`, `order.placed`, `order.paid`, `order.shipped`, `order.cancelled`.

## Запуск

```bash
cp .env.example .env        # и поменять пароли
docker compose up -d --build
curl http://localhost:8080/api/health
```

Порты на хосте: бэкенд — 8080, PostgreSQL — 5433, S3 API хранилища — 9002.

Файлы хранятся в SeaweedFS (`chrislusf/seaweedfs`, режим `mini` — все компоненты
и S3-шлюз в одном контейнере `storage`, данные в томе `storage-data`). Раньше здесь
был MinIO, поэтому переменные по-прежнему называются `MINIO_*`: бэкенд подключается
к любому S3-совместимому хранилищу. Ключи доступа задаются в `.env`
(`MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`) — при старте контейнер сам записывает из них
конфигурацию S3-пользователя, отдельный файл править не нужно. Бакет `media`
бэкенд создаёт сам при запуске.

Без Docker (нужны запущенные PostgreSQL с PostGIS и S3-хранилище;
удобно поднять только их: `docker compose up -d db storage` и указать
`MINIO_ENDPOINT=localhost:9002`):

```bash
go run ./cmd/server
```

Переменные окружения описаны в `internal/infrastructure/config/config.go`.

## Как читать код

Начните с `cmd/server/main.go` и `internal/app/wiring.go` — там видно, как собирается приложение.
Модуль удобно читать сверху вниз на примере `auth`: `domain/auth` → `service/auth` (и его тесты на
фейках) → `repository/postgres/users.go` → `transport/http/auth`.

| Файл | Что посмотреть |
| --- | --- |
| `internal/domain/errs/errs.go` | ошибки домена с родом; перевод в HTTP — в `transport/http/httpx` |
| `internal/infrastructure/postgres/tx.go` | `TxManager.Do`, транзакция в `ctx`, вложенные транзакции |
| `internal/transport/http/middleware/auth.go` | проверка токена и ролей |
| `internal/repository/postgres/queries/catalog.sql` | геопоиск: `ST_DWithin`, `ST_MakeEnvelope`, необязательные фильтры |
| `internal/domain/media/file.go` | правила файлов: типы по содержимому, панорамы, кто видит файл заявки |
| `internal/service/submission/moderation.go` | одобрение и отказ в одной транзакции, `FOR UPDATE`, событие |
| `internal/service/shop/orders.go` | заказ: блокировка остатков, снимок цен, webhook ЮKassa с перепроверкой |
| `internal/infrastructure/outbox/` | очередь событий: `FOR UPDATE SKIP LOCKED`, аренда, повторы |
| `migrations/00006_module_schemas.sql` | таблицы модулей разнесены по схемам БД |

## SQL-запросы и sqlc

В Go-коде нет SQL: запросы лежат в `internal/repository/postgres/queries/<модуль>.sql`, а Go-функции
для них (один пакет `internal/repository/postgres/sqlc`) генерирует [sqlc](https://sqlc.dev) по схеме
из миграций. Настройки — `sqlc.yaml`. Модели sqlc не выходят за `repository/postgres`: наружу
репозитории отдают сущности домена.
После изменения запросов или миграций:

```bash
go tool sqlc generate
```

Сгенерированный код хранится в репозитории, поэтому для сборки sqlc не нужен.

## Тесты

- **Домен и сервисы** — юнит-тесты без БД и HTTP: репозитории и порты заменены фейками в памяти.
  Запускаются всегда: `go test ./internal/domain/... ./internal/service/...`.
- **Репозитории, outbox, транзакции** — интеграционные тесты на настоящей PostgreSQL
  (без `TEST_DATABASE_URL` пропускаются), миграции применяют сами.
- **Обработчики** — `httptest` с фейковым сервисом: права, коды ответов, формат JSON.

```bash
docker run -d --name heritage-test-db -e POSTGRES_USER=heritage -e POSTGRES_PASSWORD=test \
  -e POSTGRES_DB=heritage -p 5434:5432 postgis/postgis:16-3.4
TEST_DATABASE_URL=postgres://heritage:test@localhost:5434/heritage?sslmode=disable go test -p 1 ./...
```

`-p 1` — пакеты по очереди, чтобы они не применяли миграции к пустой базе одновременно.

## Почта и оплата

- **Письма.** Модули пишут события в outbox, модуль notification превращает их в письма и отправляет
  через SMTP. Без `SMTP_HOST` письма только пишутся в лог бэкенда — удобно при разработке.
- **ЮKassa.** Без `YOOKASSA_SHOP_ID` заказы оформляются, но ссылки на оплату нет
  (`POST /api/orders/{id}/pay` отвечает 503). Для проверки — тестовый магазин ЮKassa.
  В личном кабинете укажите адрес уведомлений `https://<сайт>/api/payments/yookassa/webhook`.
  Уведомления ЮKassa не подписаны, поэтому бэкенд не верит телу запроса и перепроверяет
  статус платежа запросом к API. Дополнительно можно пропускать на этот адрес только IP ЮKassa в nginx.

## API

Защищённые методы требуют заголовок `Authorization: Bearer <token>` (токен из `/api/auth/login`).

| Метод | Адрес | Кто | Что делает |
| --- | --- | --- | --- |
| GET | `/api/health` | все | проверка, что сервер и БД работают |
| POST | `/api/auth/register` | все | регистрация `{email, password, name}` |
| POST | `/api/auth/login` | все | вход `{email, password}` → `{token, user}` |
| GET | `/api/auth/me` | пользователь | текущий пользователь |
| GET | `/api/admin/users` | администратор | пользователи: `limit`, `offset` |
| PUT | `/api/admin/users/{id}/role` | администратор | назначить роль `{role: user\|moderator\|admin}` |
| PUT | `/api/admin/users/{id}/blocked` | модератор | заблокировать `{blocked: true}` (модератор — только пользователей) |
| GET | `/api/categories` | все | категории объектов |
| POST | `/api/categories` | администратор | создать категорию `{slug, name}` |
| PUT | `/api/categories/{id}` | администратор | изменить категорию |
| DELETE | `/api/categories/{id}` | администратор | удалить пустую категорию |
| GET | `/api/objects` | все | список: `category`, `status`, `q`, `lat`+`lng`+`radius`, `bbox` |
| GET | `/api/objects/{slug}` | все | карточка объекта и его медиафайлы |
| POST | `/api/objects` | модератор | создать объект |
| PUT | `/api/objects/{id}` | модератор | изменить объект |
| DELETE | `/api/objects/{id}` | администратор | удалить объект |
| POST | `/api/objects/{id}/media` | модератор | загрузить файл: форма `kind`, `caption`, `yearTaken`, `sourceNote`, `source`, `projection`, `hAngle`, `author`, `license`, `file` |
| GET | `/api/media/{id}/file` | все | скачать файл (Range); файл заявки — только автор и модераторы |
| DELETE | `/api/media/{id}` | модератор | удалить файл |
| GET | `/api/routes` | все | тематические маршруты |
| GET | `/api/routes/{slug}` | все | маршрут с линией (GeoJSON) и остановками |
| POST | `/api/routes/build` | все | построить маршрут без сохранения `{objects: [slug, ...]}` |
| POST | `/api/routes` | администратор | создать тематический маршрут `{slug, title, description, objects}` |
| GET | `/api/me/routes` | пользователь | мои сохранённые маршруты |
| POST | `/api/me/routes` | пользователь | сохранить свой маршрут `{title, objects}` |
| DELETE | `/api/routes/{id}` | владелец или администратор | удалить маршрут |
| POST | `/api/submissions` | пользователь | заявка `{kind: new_object\|addition, objectSlug?, title, description, address, coordinates?, period, source}` |
| POST | `/api/submissions/{id}/media` | автор | прикрепить файл к своей заявке на модерации (до 10) |
| GET | `/api/submissions/{id}` | автор или модератор | заявка с файлами и решением |
| GET | `/api/me/submissions` | пользователь | мои заявки и их статус |
| GET | `/api/moderation/submissions` | модератор | очередь: `status` (по умолчанию `pending`), `limit`, `offset` |
| POST | `/api/moderation/submissions/{id}/approve` | модератор | одобрить; для нового объекта — поля объекта `{slug, category, protectionStatus, ...}` |
| POST | `/api/moderation/submissions/{id}/reject` | модератор | отклонить `{reason, blockAuthor}` |
| GET | `/api/products` | все | товары: `category` |
| GET | `/api/products/{slug}` | все | карточка товара |
| POST | `/api/products` | администратор | создать товар `{slug, title, description, category, priceKopecks, stock, imageUrl}` |
| PUT | `/api/products/{id}` | администратор | изменить товар |
| DELETE | `/api/products/{id}` | администратор | снять с продажи (заказы сохраняются) |
| GET | `/api/cart` | пользователь | корзина с суммой |
| PUT | `/api/cart/items/{productId}` | пользователь | количество товара `{quantity}` (0 — убрать) |
| DELETE | `/api/cart/items/{productId}` | пользователь | убрать товар |
| POST | `/api/orders` | пользователь | оформить заказ `{deliveryAddress, phone}` → заказ с `paymentUrl` |
| GET | `/api/me/orders` | пользователь | мои заказы |
| GET | `/api/orders/{id}` | владелец или администратор | заказ с позициями |
| POST | `/api/orders/{id}/pay` | владелец | ссылка на оплату (создаёт платёж, если его нет) |
| POST | `/api/orders/{id}/cancel` | владелец | отменить неоплаченный заказ |
| GET | `/api/admin/orders` | администратор | все заказы: `status`, `limit`, `offset` |
| PUT | `/api/admin/orders/{id}/status` | администратор | статус: оплачен → отправлен → выполнен, отмена |
| POST | `/api/payments/yookassa/webhook` | ЮKassa | уведомление об оплате |

Виды файлов (`kind`): `photo`, `archive_photo` (до 15 МБ, JPEG/PNG/WebP), `audio` (до 40 МБ),
`model3d` (до 50 МБ, только `.glb`).

Панорамы (`kind=panorama`, JPEG/WebP до 30 МБ): `projection` — `full_360` (полная сфера) или `partial`
(часть круга, угол обзора в градусах — `hAngle`). У каждого файла `source` — `photo` или
`ai_reconstruction`: изображения, созданные нейросетью, сайт помечает как «реконструкция ИИ».
`sourceNote` — откуда материал (архив, книга), `author` и `license` — автор и лицензия.

Деньги в API — целые копейки (`450000` = 4500 ₽), чтобы не было ошибок округления.

## Свой OSRM

Подготовка данных один раз (Южный федеральный округ с Geofabrik):

```bash
mkdir osrm-data && curl -L -o osrm-data/astrakhan.osm.pbf https://download.geofabrik.de/russia/south-fed-district-latest.osm.pbf
docker run --rm -v ./osrm-data:/data osrm/osrm-backend:v5.27.1 osrm-extract -p /opt/foot.lua /data/astrakhan.osm.pbf
docker run --rm -v ./osrm-data:/data osrm/osrm-backend:v5.27.1 osrm-partition /data/astrakhan.osrm
docker run --rm -v ./osrm-data:/data osrm/osrm-backend:v5.27.1 osrm-customize /data/astrakhan.osrm
```

Затем `OSRM_URL=http://osrm:5000` в `.env` и `docker compose --profile osrm up -d`.
