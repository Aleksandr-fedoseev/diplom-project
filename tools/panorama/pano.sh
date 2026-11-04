#!/usr/bin/env bash
# Панорама 360° по одной фотографии — целиком, одной командой (нужен запущенный ComfyUI).
#   pano.sh ФОТО ИМЯ HFOV [НАКЛОН] [SEED]
# 1) фото проецируется на развёртку 2:1 (project_photo.py);
# 2) FLUX.2 [klein] + LoRA 360 ERP дорисовывает зелёное;
# 3) шов на стыке краёв ищется автоматически и при необходимости дорисовывается заново (seam.sh);
# 4) увеличение до 4096x2048, исходное фото кладётся поверх → output/pano/ИМЯ_360.jpg
# Переменные окружения: COMFY_URL (по умолчанию http://127.0.0.1:8188), COMFY_ROOT (~/ai/ComfyUI).
set -euo pipefail
export LC_ALL=C

DIR=$(dirname "$(realpath "$0")")
COMFY=${COMFY_URL:-http://127.0.0.1:8188}
ROOT=${COMFY_ROOT:-$HOME/ai/ComfyUI}
INPUT="$ROOT/input"
OUTPUT="$ROOT/output"
SEAM_MAX=1.5    # шов есть, если стык краёв резче среднего перепада соседних столбцов в столько раз
SEAM_TRIES=3    # попыток исправить шов (каждая с новым seed)

[ $# -ge 3 ] || { echo "usage: $0 ФОТО ИМЯ HFOV [НАКЛОН] [SEED]"; exit 1; }
photo=$1 name=$2 hfov=$3 pitch=${4:-0} seed=${5:-$RANDOM}

log() { printf '[%s] %s\n' "$(date +%T)" "$*" >&2; }

curl -sf -m 5 "$COMFY/system_stats" >/dev/null || { echo "ComfyUI не отвечает на $COMFY"; exit 1; }

# Отправляет схему в ComfyUI, ждёт и печатает путь к первой сохранённой картинке.
run() {
    local wf=$1 id hist status
    id=$(jq -c '{prompt: .}' <<<"$wf" | curl -sf -X POST -H 'Content-Type: application/json' -d @- "$COMFY/prompt" | jq -r .prompt_id)
    for _ in $(seq 1800); do
        hist=$(curl -sf "$COMFY/history/$id" | jq -c --arg id "$id" '.[$id] // empty')
        if [ -n "$hist" ]; then
            status=$(jq -r '.status.status_str' <<<"$hist")
            [ "$status" = success ] || { echo "ComfyUI: задача $id завершилась с ошибкой" >&2; jq '.status.messages' <<<"$hist" >&2; return 1; }
            jq -r --arg out "$OUTPUT" '[.outputs[].images[]?][0] | "\($out)/\(.subfolder)/\(.filename)"' <<<"$hist"
            return
        fi
        sleep 1
    done
    echo "ComfyUI: задача $id не завершилась за 30 мин" >&2; return 1
}

# Генерация: схема klein без увеличения, на входе картинка из input/.
generate() {
    jq --arg img "$1" --arg prefix "$2" --argjson seed "$3" '
        del(.["14"], .["15"], .["16"], .["17"])
        | .["5"].inputs.image = $img
        | .["11"].inputs.seed = $seed
        | .["13"].inputs.filename_prefix = $prefix' "$DIR/klein_panorama_from_photo.json" | run "$(cat)"
}

# Во сколько раз стык левого и правого краёв резче среднего перепада между соседними столбцами.
seam_ratio() {
    local f=$1 w h s r
    read -r w h < <(magick identify -format '%w %h\n' "$f")
    mae() { magick compare -metric MAE "$1" "$2" null: 2>&1 | sed -n 's/.*(\(.*\)).*/\1/p'; }
    s=$(mae <(magick "$f" -crop 1x${h}+0+0 +repage png:-) <(magick "$f" -crop 1x${h}+$((w - 1))+0 +repage png:-))
    r=$(mae <(magick "$f" -crop $((w - 1))x${h}+0+0 +repage png:-) <(magick "$f" -crop $((w - 1))x${h}+1+0 +repage png:-))
    awk -v s="$s" -v r="$r" 'BEGIN { printf "%.2f", s / r }'
}
has_seam() { awk -v x="$1" -v m="$SEAM_MAX" 'BEGIN { exit !(x > m) }'; }

log "1/4 проекция фото (HFOV $hfov°, наклон $pitch°)"
uv run -q --with numpy --with pillow "$DIR/project_photo.py" "$photo" "$name" "$hfov" "$pitch" "$INPUT" >&2

log "2/4 генерация панорамы (seed $seed)"
pano=$(generate "control_$name.png" "pano/${name}_raw" "$seed")
ratio=$(seam_ratio "$pano")
log "    стык краёв: $ratio (порог $SEAM_MAX)"

try=0
while has_seam "$ratio" && [ $try -lt $SEAM_TRIES ]; do
    try=$((try + 1))
    log "3/4 исправление шва, попытка $try"
    "$DIR/seam.sh" prep "$pano" "$name" >/dev/null
    patch=$(generate "seam_$name.png" "pano/${name}_seamfix" $((seed + try)))
    fixed="$OUTPUT/pano/${name}_fixed_$try.png"
    "$DIR/seam.sh" merge "$pano" "$patch" "$fixed" >/dev/null
    new=$(seam_ratio "$fixed")
    log "    стык краёв: $new"
    if awk -v a="$new" -v b="$ratio" 'BEGIN { exit !(a < b) }'; then
        pano=$fixed ratio=$new
    fi
done
has_seam "$ratio" && log "    ВНИМАНИЕ: шов остался ($ratio), проверьте панораму вручную"
[ $try -eq 0 ] && log "3/4 шва нет"

log "4/4 увеличение до 4096x2048 и наложение фото"
cp "$pano" "$INPUT/final_$name.png"
big=$(jq --arg img "final_$name.png" --arg prefix "pano/${name}_4k" '
    .["1"].inputs.image = $img | .["5"].inputs.filename_prefix = $prefix' "$DIR/upscale_only.json" | run "$(cat)")
out="$OUTPUT/pano/${name}_360.jpg"
magick "$big" "$INPUT/overlay_$name.png" -composite -quality 88 "$out"
log "готово: $out"
echo "$out"
