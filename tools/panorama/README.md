# Панорама 360° по одной фотографии

Инструмент модератора: по обычной фотографии здания строит сферическую панораму 4096×2048
(равнопромежуточная проекция 2:1). Здание остаётся подлинным — это исходный снимок,
окружение дорисовывает нейросеть FLUX.2 [klein] с LoRA для достраивания панорам 360°.
Готовый файл загружается на платформу как медиафайл объекта с пометкой «реконструкция ИИ».

Работает на ПК модератора с видеокартой NVIDIA от 8 ГБ; серверу платформы видеокарта не нужна.

## Запуск

```bash
tools/panorama/pano.sh ФОТО ИМЯ HFOV [НАКЛОН] [SEED]
# пример
tools/panorama/pano.sh 266.jpg kononov 95 6
```

- `HFOV` — горизонтальный угол обзора снимка, градусы (телефон — около 70, широкий угол — 90–100);
- `НАКЛОН` — на сколько градусов камера поднята вверх (обычно 0–8);
- `SEED` — зерно генерации; без него берётся случайное, повторный запуск даёт другой вариант.

Результат: `~/ai/ComfyUI/output/pano/ИМЯ_360.jpg`. Одна панорама — около 6 мин, исправление шва — ещё около 5 мин
(RTX 5050 8 ГБ).

## Что происходит внутри

| Шаг | Файл | Что делает |
|---|---|---|
| 1 | `project_photo.py` | проецирует фото на развёртку по модели камеры-обскуры; остальное закрашивает зелёным (`control_ИМЯ.png`), отдельно готовит фото с прозрачностью для наложения (`overlay_ИМЯ.png`); небо на фото находится автоматически и в наложение не попадает |
| 2 | `klein_panorama_from_photo.json` | FLUX.2 [klein] + LoRA дорисовывает зелёную область (2048×1024) |
| 3 | `seam.sh` | шов на стыке левого и правого краёв: если стык резче среднего перепада соседних столбцов больше чем в 1,5 раза, панорама сдвигается на половину ширины, полоса шва дорисовывается заново и вклеивается обратно (до 3 попыток) |
| 4 | `upscale_only.json` | увеличение RealESRGAN до 4096×2048, поверх кладётся исходное фото |

Схемы `*.json` — в формате API ComfyUI; `pano.sh` отправляет их на `POST /prompt` и ждёт результата через `GET /history/{id}`.

## Что нужно установить

- [ComfyUI](https://github.com/comfyanonymous/ComfyUI), запущенный на `http://127.0.0.1:8188`;
- `jq`, `curl`, ImageMagick 7 (`magick`), [uv](https://docs.astral.sh/uv/) (сам ставит numpy и pillow для `project_photo.py`).

Модели — в папки `ComfyUI/models/…`:

| Папка | Файл | Где взять |
|---|---|---|
| `diffusion_models` | `flux-2-klein-base-4b-fp8.safetensors` | [black-forest-labs/FLUX.2-klein-base-4b-fp8](https://huggingface.co/black-forest-labs/FLUX.2-klein-base-4b-fp8) |
| `text_encoders` | `qwen_3_4b_fp4_flux2.safetensors` | [Comfy-Org/Flux-2-Klein-4B](https://huggingface.co/Comfy-Org/Flux-2-Klein-4B), папка `split_files/text_encoders` |
| `vae` | `flux2-vae.safetensors` | там же, `split_files/vae` |
| `loras` | `flux-2-klein-4B-360-erp-outpaint-lora_V1.safetensors` | [nomadoor/flux-2-klein-4B-360-erp-outpaint-lora](https://huggingface.co/nomadoor/flux-2-klein-4B-360-erp-outpaint-lora) |
| `upscale_models` | `RealESRGAN_x4plus.pth` | [Real-ESRGAN v0.1.0](https://github.com/xinntao/Real-ESRGAN/releases/download/v0.1.0/RealESRGAN_x4plus.pth) |

FLUX.2 [klein] 4B и LoRA распространяются под Apache 2.0 (коммерческое использование разрешено).

Переменные окружения: `COMFY_URL` — адрес ComfyUI, `COMFY_ROOT` — папка ComfyUI (по умолчанию `~/ai/ComfyUI`).

## Загрузка на платформу

`POST /api/objects/{id}/media` (роль: модератор), форма `multipart/form-data`:

```bash
curl -H "Authorization: Bearer $TOKEN" \
  -F file=@kononov_360.jpg -F kind=panorama -F projection=full_360 \
  -F source=ai_reconstruction -F caption="Вид на главный фасад" \
  -F author="автор исходного фото" -F license="…" \
  http://localhost:8080/api/objects/$OBJECT_ID/media
```

Сайт показывает такие панорамы со значком «Реконструкция ИИ» и пояснением, что окружение дорисовано.
Права на исходную фотографию принадлежат автору съёмки — для публикации нужно его разрешение.
