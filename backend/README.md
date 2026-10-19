# Бэкенд платформы наследия (Go, монолит)

Один Go-сервис по диаграмме развёртывания (`Diagrams/Deployment.png`): модули авторизации,
каталога и маршрутов, медиафайлов. Данные — PostgreSQL 16 + PostGIS, файлы — MinIO,
маршруты по улицам — OSRM.

Сделано сейчас: авторизация и роли, каталог объектов с геопоиском, маршруты через OSRM,
загрузка фото, аудио и 3D. Дальше по плану: заявки и модерация, магазин и оплата ЮKassa,
уведомления по email.

## Запуск

```bash
cp .env.example .env        # и поменять пароли
docker compose up -d --build
curl http://localhost:8080/api/health
```

Без Docker (нужны запущенные PostgreSQL с PostGIS и MinIO):

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
| `internal/catalog/repository.go` | SQL с динамическими фильтрами, `ST_DWithin`, `ST_MakeEnvelope` |
| `internal/media/service.go` | потоковая загрузка в MinIO, проверка типа по содержимому, Range-запросы |
| `internal/osrm/client.go` | HTTP-клиент к внешнему сервису с таймаутом |
| `internal/routes/service.go` | транзакция: маршрут и его точки сохраняются вместе |

Слои в каждом модуле: `handler.go` разбирает HTTP-запрос и пишет ответ,
`service.go` / `repository.go` — логика и SQL. Обработчики не знают про SQL, а SQL не знает про HTTP.

## API

Защищённые методы требуют заголовок `Authorization: Bearer <token>` (токен из `/api/auth/login`).

| Метод | Адрес | Кто | Что делает |
| --- | --- | --- | --- |
| GET | `/api/health` | все | проверка, что сервер и БД работают |
| POST | `/api/auth/register` | все | регистрация `{email, password, name}` |
| POST | `/api/auth/login` | все | вход `{email, password}` → `{token, user}` |
| GET | `/api/auth/me` | пользователь | текущий пользователь |
| GET | `/api/categories` | все | категории объектов |
| GET | `/api/objects` | все | список: `category`, `status`, `q`, `lat`+`lng`+`radius`, `bbox` |
| GET | `/api/objects/{slug}` | все | карточка объекта и его медиафайлы |
| POST | `/api/objects` | модератор | создать объект |
| PUT | `/api/objects/{id}` | модератор | изменить объект |
| DELETE | `/api/objects/{id}` | администратор | удалить объект |
| POST | `/api/objects/{id}/media` | модератор | загрузить файл: форма `kind`, `caption`, `file` |
| GET | `/api/media/{id}/file` | все | скачать файл (поддерживает Range) |
| DELETE | `/api/media/{id}` | модератор | удалить файл |
| GET | `/api/routes` | все | тематические маршруты |
| GET | `/api/routes/{slug}` | все | маршрут с линией (GeoJSON) и остановками |
| POST | `/api/routes/build` | все | построить маршрут без сохранения `{objects: [slug, ...]}` |
| POST | `/api/routes` | администратор | создать тематический маршрут `{slug, title, description, objects}` |
| GET | `/api/me/routes` | пользователь | мои сохранённые маршруты |
| POST | `/api/me/routes` | пользователь | сохранить свой маршрут `{title, objects}` |
| DELETE | `/api/routes/{id}` | владелец или администратор | удалить маршрут |

Виды файлов (`kind`): `photo`, `archive_photo` (до 15 МБ, JPEG/PNG/WebP), `audio` (до 40 МБ),
`model3d` (до 50 МБ, только `.glb`).

## Свой OSRM

Подготовка данных один раз (Южный федеральный округ с Geofabrik):

```bash
mkdir osrm-data && curl -L -o osrm-data/astrakhan.osm.pbf https://download.geofabrik.de/russia/south-fed-district-latest.osm.pbf
docker run --rm -v ./osrm-data:/data osrm/osrm-backend:v5.27.1 osrm-extract -p /opt/foot.lua /data/astrakhan.osm.pbf
docker run --rm -v ./osrm-data:/data osrm/osrm-backend:v5.27.1 osrm-partition /data/astrakhan.osrm
docker run --rm -v ./osrm-data:/data osrm/osrm-backend:v5.27.1 osrm-customize /data/astrakhan.osrm
```

Затем `OSRM_URL=http://osrm:5000` в `.env` и `docker compose --profile osrm up -d`.
