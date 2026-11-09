#!/usr/bin/env bash
# Наполняет запущенный бэкенд стартовыми данными через API:
#   1) фотографии и панорамы объектов из public/heritage — в S3-хранилище;
#   2) тематические маршруты (линии по улицам строит OSRM);
#   3) демо-пользователи для проверки ролей: user@demo.test и moderator@demo.test.
# Повторный запуск безопасен: что уже есть, пропускается.
#
#   tools/seed/seed.sh                    # API на http://localhost:8080/api
#   API=http://сервер/api tools/seed/seed.sh
#
# Вход администратора берётся из backend/.env (ADMIN_EMAIL, ADMIN_PASSWORD)
# или из переменных окружения с теми же именами.
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
API=${API:-http://localhost:8080/api}
DEMO_PASSWORD=${DEMO_PASSWORD:-demo-password}

for cmd in curl jq; do
  command -v "$cmd" >/dev/null || { echo "нужна утилита $cmd" >&2; exit 1; }
done

if [[ -f "$ROOT/backend/.env" ]]; then
  ADMIN_EMAIL=${ADMIN_EMAIL:-$(grep -E '^ADMIN_EMAIL=' "$ROOT/backend/.env" | cut -d= -f2-)}
  ADMIN_PASSWORD=${ADMIN_PASSWORD:-$(grep -E '^ADMIN_PASSWORD=' "$ROOT/backend/.env" | cut -d= -f2-)}
fi
: "${ADMIN_EMAIL:?не задан ADMIN_EMAIL}" "${ADMIN_PASSWORD:?не задан ADMIN_PASSWORD}"

# api МЕТОД ПУТЬ [аргументы curl] — печатает тело ответа, при ошибке HTTP завершает скрипт.
api() {
  local method=$1 path=$2 out code
  shift 2
  out=$(mktemp)
  code=$(curl -sS -o "$out" -w '%{http_code}' -X "$method" "$API$path" \
    ${TOKEN:+-H "Authorization: Bearer $TOKEN"} "$@")
  if [[ $code -ge 400 ]]; then
    echo "ошибка $code: $method $path — $(cat "$out")" >&2
    rm -f "$out"
    return 1
  fi
  cat "$out"
  rm -f "$out"
}

json() { api "$1" "$2" -H 'Content-Type: application/json' -d "$3"; }

login() {
  json POST /auth/login "$(jq -n --arg e "$1" --arg p "$2" '{email: $e, password: $p}')" | jq -r .token
}

api GET /health >/dev/null || { echo "бэкенд не отвечает: $API" >&2; exit 1; }
TOKEN=$(login "$ADMIN_EMAIL" "$ADMIN_PASSWORD")
echo "вход администратора: $ADMIN_EMAIL"

# --- 1. Медиафайлы объектов ---------------------------------------------------
# upload ID ФАЙЛ ВИД [поля формы...]
upload() {
  local id=$1 file=$2 kind=$3
  shift 3
  api POST "/objects/$id/media" -F "kind=$kind" -F "file=@$file;type=image/jpeg" "$@" >/dev/null
}

for dir in "$ROOT"/public/heritage/*/; do
  slug=$(basename "$dir")
  card=$(api GET "/objects/$slug" 2>/dev/null) || { echo "  $slug: нет в каталоге, пропуск"; continue; }
  id=$(jq -r .object.id <<<"$card")
  if [[ $(jq '.media | length' <<<"$card") -gt 0 ]]; then
    echo "  $slug: файлы уже загружены"
    continue
  fi
  [[ -f $dir/cover.jpg ]] && upload "$id" "$dir/cover.jpg" photo -F source=photo -F "caption=Главный фасад"
  # Панорамы дома Кононова созданы нейросетью по фотографии — помечаются «реконструкция ИИ».
  [[ -f $dir/panorama-facade.jpg ]] && upload "$id" "$dir/panorama-facade.jpg" panorama \
    -F projection=full_360 -F source=ai_reconstruction -F "caption=Вид на главный фасад"
  [[ -f $dir/panorama-street.jpg ]] && upload "$id" "$dir/panorama-street.jpg" panorama \
    -F projection=full_360 -F source=ai_reconstruction -F "caption=Вид с улицы"
  echo "  $slug: загружено"
done

# --- 2. Тематические маршруты -------------------------------------------------
route() {
  local slug=$1 title=$2 description=$3
  shift 3
  if api GET "/routes/$slug" >/dev/null 2>&1; then
    echo "  маршрут $slug уже есть"
    return
  fi
  json POST /routes "$(jq -n --arg s "$slug" --arg t "$title" --arg d "$description" \
    '{slug: $s, title: $t, description: $d, objects: $ARGS.positional}' --args "$@")" >/dev/null
  echo "  маршрут $slug создан"
}

route nikolskaya-ulitsa 'Главная купеческая улица: Никольская и Биржа' \
  'Пешеходный маршрут по парадной улице: здание Биржи, банк Лидваля, дома Агамжановой и Тавризова.' \
  birzha azovo-donskoy-bank dom-agamzhanovoy dom-tavrizova
route usadby-reki-kutum 'Купеческие особняки набережной Кутума' \
  'Осмотр усадеб Губина, Шелехова, Будагова и первой городской электростанции.' \
  dom-gubina elektrostantsiya dom-budagova dom-shelehova

# --- 3. Демо-пользователи для проверки ролей ---------------------------------
demo_user() {
  local email=$1 name=$2 role=$3 id
  TOKEN='' json POST /auth/register \
    "$(jq -n --arg e "$email" --arg p "$DEMO_PASSWORD" --arg n "$name" '{email: $e, password: $p, name: $n}')" \
    >/dev/null 2>&1 || true # уже зарегистрирован
  id=$(api GET "/admin/users?limit=200" | jq -r --arg e "$email" '.[] | select(.email == $e) | .id')
  [[ -n $id ]] || { echo "  $email: не найден" >&2; return 1; }
  if [[ $role != user ]]; then
    json PUT "/admin/users/$id/role" "{\"role\": \"$role\"}" >/dev/null
  fi
  echo "  $email ($role), пароль: $DEMO_PASSWORD"
}

demo_user user@demo.test 'Демо-пользователь' user
demo_user moderator@demo.test 'Демо-модератор' moderator

echo "готово"
