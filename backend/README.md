# Бэкенд платформы наследия (Go, монолит)

Один Go-сервис по диаграмме развёртывания (`Diagrams/Deployment.png`). Данные — PostgreSQL 16 + PostGIS,
файлы — S3-хранилище (SeaweedFS, клиент minio-go), маршруты по улицам — OSRM, оплата — ЮKassa,
письма — SMTP.

Модули (`internal/`):

| Модуль | Что делает |
| --- | --- |
| `auth` | регистрация и вход, JWT, роли, блокировка, управление пользователями |
| `catalog` | объекты и категории, геопоиск (PostGIS); владелец таблиц объектов |
| `routes` | тематические и пользовательские маршруты, линия по улицам через OSRM |
| `media` | фото, аудио и 3D в S3-хранилище; файлы объектов и заявок |
| `submissions` | заявки пользователей и модерация: одобрение с публикацией, отказ с причиной |
| `shop` | товары, корзина, заказы, оплата через ЮKassa и её уведомления (webhook) |
| `notifications` | очередь писем в БД и фоновая отправка с повторными попытками |

Архитектура — модульный монолит: модули поделены по предметным областям, внутри каждого слои
`handler.go` (HTTP) → `service.go` (логика) → `store/` (доступ к БД, код генерирует sqlc).
Модуль работает только со своими таблицами, а сведения из других модулей получает через их
сервисы — небольшие интерфейсы объявлены в модуле-потребителе (например, `routes.Places`,
`media.Objects`). Все модули собираются в `cmd/server/main.go`.

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

Переменные окружения описаны в `internal/config/config.go`.

## Как читать код

Начните с `cmd/server/main.go` — там видно, как собирается приложение. Дальше по порядку:

| Файл | Что посмотреть |
| --- | --- |
| `internal/config/config.go` | настройки из переменных окружения, проверка обязательных |
| `internal/database/migrations/*.sql` | схема БД; PostGIS-тип `geography`, GIST-индекс |
| `internal/database/database.go` | пул соединений pgx, миграции goose, `go:embed` |
| `internal/httpx/httpx.go` | обработчик, возвращающий `error`; ошибки для клиента и для лога |
| `internal/server/server.go` | маршрутизатор chi, цепочка middleware |
| `internal/auth/` | bcrypt, JWT, middleware `Authenticate` и `RequireRole` |
| `internal/catalog/queries.sql` | геопоиск: `ST_DWithin`, `ST_MakeEnvelope`, необязательные фильтры через `sqlc.narg` |
| `internal/media/service.go` | потоковая загрузка в S3-хранилище, проверка типа по содержимому, Range-запросы |
| `internal/osrm/client.go` | HTTP-клиент к внешнему сервису с таймаутом |
| `internal/routes/service.go` | транзакция: маршрут и его точки сохраняются вместе; остановки — у каталога |
| `internal/submissions/service.go` | одобрение в одной транзакции: объект, файлы, решение, письмо; `FOR UPDATE` |
| `internal/shop/service.go` | оформление заказа: блокировка остатков, снимок цен; webhook ЮKassa с перепроверкой |
| `internal/notifications/service.go` | очередь писем: `FOR UPDATE SKIP LOCKED`, аренда, повтор через 1, 4, 9, 16 минут |

## SQL-запросы и sqlc

В Go-коде нет SQL: запросы каждого модуля лежат в его `queries.sql`, а Go-функции для них
(пакет `store` модуля) генерирует [sqlc](https://sqlc.dev) по схеме из миграций. Настройки — `sqlc.yaml`.
После изменения запросов или миграций:

```bash
go tool sqlc generate
```

Сгенерированный код хранится в репозитории, поэтому для сборки sqlc не нужен.

## Тесты

Интеграционные тесты работают с настоящей PostgreSQL (без неё пропускаются), миграции применяют сами.
S3-хранилище и OSRM в тестах заменены поддельными реализациями.

```bash
docker run -d --name heritage-test-db -e POSTGRES_USER=heritage -e POSTGRES_PASSWORD=test \
  -e POSTGRES_DB=heritage -p 5434:5432 postgis/postgis:16-3.4
TEST_DATABASE_URL=postgres://heritage:test@localhost:5434/heritage?sslmode=disable go test -p 1 ./...
```

`-p 1` — пакеты по очереди, чтобы они не применяли миграции к пустой базе одновременно.

## Почта и оплата

- **Письма.** Модули ставят письмо в очередь (`email_jobs`), фоновый обработчик отправляет его
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
| POST | `/api/objects/{id}/media` | модератор | загрузить файл: форма `kind`, `caption`, `yearTaken`, `source`, `file` |
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
