# Обучение LoRA и перенос проекта на другой ПК

Актуально на 30.10.2026. Онлайн-версия: https://claude.ai/artifact/KCsSDCVXUmNmpdfi1Xmzgw

## Задача для агента

Цель — обучить LoRA для SDXL на фото астраханской архитектуры, чтобы генерировать реконструкции и панорамы для платформы наследия (ВКР, см. `Docs/README.md` и `backend/README.md`).

1. Проверить ПК по разделу «Подготовка ПК»: `nvidia-smi` (RTX 5050, 8 ГБ, драйвер 570+), Python 3.11, git, ~30 ГБ на диске.
2. Установить kohya sd-scripts и скачать SDXL по разделу «Установка kohya sd-scripts и SDXL»; убедиться, что `torch.cuda.is_available()` возвращает `True`.
3. Датасет в репозиторий не входит: его переносят вручную (архив `lora-dataset.tar.gz`, 177 МБ). Распаковать в `~/lora-dataset/` и проверить структуру по таблице в разделе «Что перенести на другой ПК».
4. Сначала обучить LoRA стиля (`astrakhan_style`) командой из раздела «Обучение под 8 ГБ», затем при необходимости — LoRA отдельных зданий.
5. Проверить результат в ComfyUI по разделу «Проверка результата в ComfyUI»: пары картинок с одним seed без LoRA и с LoRA, выбор лучшей эпохи.
6. При ошибках — таблица «Частые ошибки». Обученные файлы `.safetensors` в git не добавлять (по 100–200 МБ).


## Linux или Windows

Обучать можно на обеих системах: программы обучения (kohya sd-scripts, OneTrainer) и ComfyUI работают и там, и там. Выбирайте ту систему, которая уже стоит на ПК с видеокартой.

|  | Linux | Windows |
| --- | --- | --- |
| Драйвер NVIDIA | ставится из репозитория (на CachyOS/Arch — пакет `nvidia-open`) | установщик с сайта NVIDIA |
| Скорость обучения | немного выше, система тратит меньше видеопамяти | чуть ниже: рабочий стол Windows занимает 0,5–1 ГБ видеопамяти |
| Установка программ | через терминал, по шагам ниже | так же, команды почти те же |
| Риски | нужно следить, чтобы драйвер и CUDA были свежими | реже проблемы с драйвером |

Главное условие для RTX 5050 (поколение Blackwell): **драйвер NVIDIA версии 570 или новее и PyTorch, собранный с CUDA 12.8 или новее**. Со старыми сборками PyTorch карта не запустится, ошибка будет вида «no kernel image is available».

Если на втором ПК стоит Windows и вы не хотите ничего переустанавливать — обучайте на Windows. Если стоит Linux — на Linux, на 8 ГБ видеопамяти это даже немного выгоднее.

## Что перенести на другой ПК

Проект готов к переносу: весь код лежит на GitHub, локальная копия с ним совпадает. Вручную нужно перенести только датасет — он не в репозитории.

| Что | Где сейчас | Размер | Как перенести |
| --- | --- | --- | --- |
| Код проекта | GitHub: `Aleksandr-fedoseev/diplom-project` | — | `git clone https://github.com/Aleksandr-fedoseev/diplom-project.git` |
| Датасет для LoRA | `~/Загрузки/lora-dataset/` | 177 МБ | флешка, облако или архив: `tar czf lora-dataset.tar.gz -C ~/Загрузки lora-dataset` |
| Оригиналы фото (необязательно) | `~/Загрузки/25.08.30 Здания и улицы/` | 9,9 ГБ | нужны, только если захотите пересобрать датасет |
| Секреты бэкенда | файла ещё нет | — | на новом ПК: `cp backend/.env.example backend/.env` и задать пароли |

Структура датасета (число в начале имени папки — сколько раз каждое фото повторяется за эпоху):

| Папка | Фото | Триггер-слово | Что даст LoRA |
| --- | --- | --- | --- |
| `style/2_astrakhan_style` | 226 | `astrakhan_style` | стиль астраханской архитектуры |
| `buildings/birzha/10_astr_birzha` | 22 | `astr_birzha` | здание Биржи |
| `buildings/tavrizov/10_astr_tavrizov` | 25 | `astr_tavrizov` | дом Тавризова |
| `buildings/kononov/10_astr_kononov` | 22 | `astr_kononov` | дом Кононова |
| `buildings/tinaki/10_astr_tinaki_church` | 22 | `astr_tinaki_church` | деревянная церковь в Тинаки |

Рядом с каждым фото лежит `.txt` с подписью, в `manifest.csv` — из какого исходного кадра сделано каждое фото. Фото уменьшены до 1344 px по длинной стороне. Права на снимки принадлежат автору съёмки — перед использованием в проекте нужно его разрешение.

## Подготовка ПК

Нужны четыре вещи: свежий драйвер NVIDIA, Python 3.11, git и около 30 ГБ свободного места (базовая модель 7 ГБ, программы с библиотеками 10–15 ГБ, кэш и результаты).

**Linux (Arch / CachyOS):**

1. Драйвер: `sudo pacman -S nvidia-open nvidia-utils`, затем перезагрузка.
2. Проверка: `nvidia-smi` — должны быть видны RTX 5050, 8 ГБ и версия драйвера 570 или выше.
3. Python и git: `sudo pacman -S git uv`, затем `uv python install 3.11`. Системный Python в Arch слишком новый для части библиотек, поэтому 3.11 ставим отдельно.

**Windows:**

1. Драйвер: скачать с nvidia.com (Game Ready или Studio, версия 570+).
2. Проверка: в PowerShell `nvidia-smi`.
3. Python 3.11 с python.org (при установке отметить «Add python.exe to PATH») и Git for Windows с git-scm.com.
4. Все команды ниже выполнять в PowerShell; вместо `source venv/bin/activate` — `venv\Scripts\activate`.

Отдельно ставить CUDA Toolkit **не нужно**: нужные библиотеки CUDA придут вместе с PyTorch.

## Установка kohya sd-scripts и SDXL

Используем [kohya-ss/sd-scripts](https://github.com/kohya-ss/sd-scripts) — самый распространённый набор скриптов для обучения LoRA. Графические оболочки (kohya\_ss GUI, OneTrainer) внутри делают то же самое; команда в терминале удобнее тем, что её легко повторить и показать в ВКР.

1. Скачать скрипты и создать окружение:

```bash
git clone https://github.com/kohya-ss/sd-scripts.git
cd sd-scripts
uv venv --python 3.11 venv
source venv/bin/activate          # Windows: venv\Scripts\activate
```

2. Поставить PyTorch с CUDA 12.8 (именно cu128 — иначе RTX 5050 не заработает) и остальные библиотеки:

```bash
uv pip install torch torchvision --index-url https://download.pytorch.org/whl/cu128
uv pip install -r requirements.txt
uv pip install bitsandbytes
```

3. Проверить, что PyTorch видит карту (должно напечатать `True` и название карты):

```bash
python -c "import torch; print(torch.cuda.is_available(), torch.cuda.get_device_name(0))"
```

4. Настроить accelerate один раз: `accelerate config default --mixed_precision bf16`.
5. Скачать базовую модель `sd_xl_base_1.0.safetensors` (около 7 ГБ) с страницы [stabilityai/stable-diffusion-xl-base-1.0](https://huggingface.co/stabilityai/stable-diffusion-xl-base-1.0) и положить, например, в `~/models/`. Лицензия модели (CreativeML Open RAIL++-M) разрешает коммерческое использование.

Если не хотите ставить `uv`, те же шаги делаются обычными `python3.11 -m venv venv` и `pip install …`. Версии библиотек меняются — если что-то не ставится, сверьтесь с README репозитория sd-scripts.

## Обучение под 8 ГБ

Начните с LoRA стиля: это один запуск на 1,5–3,5 часа, и по нему сразу видно, работает ли вся схема. Настройки ниже собраны так, чтобы SDXL поместился в 8 ГБ видеопамяти.

```bash
accelerate launch --num_cpu_threads_per_process 4 sdxl_train_network.py \
  --pretrained_model_name_or_path ~/models/sd_xl_base_1.0.safetensors \
  --train_data_dir ~/lora-dataset/style \
  --output_dir ~/lora-output --output_name astrakhan_style \
  --caption_extension .txt \
  --resolution 1024,1024 --enable_bucket --min_bucket_reso 512 --max_bucket_reso 1536 \
  --network_module networks.lora --network_dim 32 --network_alpha 16 \
  --network_train_unet_only --cache_text_encoder_outputs \
  --cache_latents --cache_latents_to_disk --no_half_vae \
  --train_batch_size 1 --max_train_epochs 8 --save_every_n_epochs 2 \
  --learning_rate 1e-4 --lr_scheduler cosine --lr_warmup_steps 100 \
  --optimizer_type AdamW8bit \
  --mixed_precision bf16 --save_precision fp16 --save_model_as safetensors \
  --gradient_checkpointing --sdpa --seed 42
```

В PowerShell перенос строки — обратный апостроф `` ` `` вместо `\`, а пути пишутся полностью (`C:\models\…`).

Что здесь экономит память:

- `--network_train_unet_only` + `--cache_text_encoder_outputs` — текстовые кодировщики не обучаются и не держатся в памяти; триггер-слово всё равно работает.
- `--cache_latents_to_disk` — фото кодируются один раз и берутся с диска.
- `--gradient_checkpointing`, `AdamW8bit`, `bf16` — стандартный набор для карт на 8–12 ГБ.

Сколько шагов и времени (время примерное, из расчёта 1,5–3,5 с на шаг; точную скорость покажет первая минута обучения):

| LoRA | Фото × повторы | Эпох | Шагов | Время |
| --- | --- | --- | --- | --- |
| Стиль (`style`) | 226 × 2 | 8 | ≈ 3600 | 1,5–3,5 ч |
| Одно здание (`buildings/<имя>`) | 22 × 10 | 10 | ≈ 2200 | 1–2 ч |

Для LoRA здания поменяйте три параметра: `--train_data_dir ~/lora-dataset/buildings/birzha`, `--output_name astr_birzha`, и `--max_train_epochs 10`.

Результат — файлы `~/lora-output/astrakhan_style-000002.safetensors`, `-000004`, `-000006` и итоговый `astrakhan_style.safetensors` (по 100–200 МБ). Промежуточные нужны, чтобы выбрать лучшую эпоху: последняя не всегда лучшая.

## Проверка результата в ComfyUI

Хорошая LoRA даёт узнаваемо астраханскую архитектуру по триггер-слову и не ломает картинку при силе 0,7–1,0. Проверяем в [ComfyUI](https://github.com/comfyanonymous/ComfyUI).

1. Установить ComfyUI в отдельное окружение тем же способом: `git clone`, `uv venv --python 3.11`, PyTorch cu128, `uv pip install -r requirements.txt`, запуск `python main.py` → браузер на `http://127.0.0.1:8188`.
2. Положить файлы: `sd_xl_base_1.0.safetensors` → `ComfyUI/models/checkpoints/`, свои LoRA → `ComfyUI/models/loras/`.
3. В стандартной схеме между узлом Load Checkpoint и текстовыми полями добавить узел **Load LoRA**, выбрать файл и силу (strength) 0,8.
4. Размер 1024×1024 или 1216×832, 25–30 шагов, CFG 6–7.

Промпты для проверки (писать по-английски, как подписи в датасете):

- `astrakhan_style, photo, two-story red brick merchant house with cast iron balcony, quiet street, summer`
- `astrakhan_style, photo, wooden house with carved gable, 19th century, sunset`
- `astr_birzha, photo, white and turquoise eclectic exchange building, tall tower, view from the square`

Как сравнивать:

- один и тот же seed без LoRA и с LoRA — эта пара картинок и есть наглядный результат для защиты;
- файлы разных эпох (`-000004`, `-000006`, итоговый) с одним seed;
- признак переобучения — картинки почти копируют фото из датасета или идут пятнами: берите более раннюю эпоху или уменьшите силу до 0,6;
- признак недообучения — разницы с картинкой без LoRA почти нет: добавьте 2–4 эпохи.

## Панорамы 360°

Панорама получается из двух LoRA сразу: ваша (стиль или здание) даёт архитектуру, готовая 360°-LoRA — правильную сферическую проекцию. Обучать панорамную LoRA самим не нужно.

1. Найти панорамную LoRA для SDXL на civitai.com или huggingface.co по запросу «360 panorama» / «equirectangular». Проверить лицензию: разрешено ли коммерческое использование и нужно ли указывать автора.
2. В ComfyUI поставить два узла Load LoRA подряд: панорамная (сила 0,8–1,0) и ваша (0,6–0,8).
3. Размер строго 2:1. На 8 ГБ надёжнее 1536×768 и потом увеличение до 3072×1536; сразу 2048×1024 тоже возможно, если включить VAE Decode (Tiled).
4. Промпт: слово-триггер панорамной LoRA (указано на её странице) + `astrakhan_style, 360 degree panorama, historic street with red brick merchant houses, cast iron balconies, summer`.
5. Проверить шов: левый и правый края должны стыковаться. Если виден стык — сдвинуть картинку на половину ширины и залатать центр инпейнтом.
6. Посмотреть панораму в просмотрщике (Photo Sphere Viewer или Pannellum) — искажения у полюсов видны только там.

Загружать результат на сайт через модератора: `POST /api/objects/{id}/media` с полями `kind=panorama`, `projection=full_360`, `source=ai_reconstruction`. По полю `source` фронтенд должен помечать такие изображения как «реконструкция ИИ» (в бэкенде это уже есть, на страницах сайта — ещё нет): модель дорисовывает то, чего не было на фото, и выдавать это за реальный вид нельзя.

## Частые ошибки

| Что видно | Причина | Что сделать |
| --- | --- | --- |
| `no kernel image is available for execution on the device` | PyTorch собран без поддержки RTX 50-й серии | переустановить torch с `--index-url …/whl/cu128` |
| `torch.cuda.is_available()` → `False` | драйвер не загружен или старый | `nvidia-smi`; драйвер 570+, перезагрузка |
| `CUDA out of memory` | не хватает 8 ГБ | по очереди: `--max_bucket_reso 1024`, `--network_dim 16`, `--fp8_base`; закрыть браузер и игры |
| Ошибка про bitsandbytes или AdamW8bit | библиотека не видит CUDA | `uv pip install -U bitsandbytes`; временно — `--optimizer_type AdamW` (памяти уйдёт больше) |
| Потеря (loss) `nan`, чёрные картинки | VAE SDXL в fp16 | оставить `--no_half_vae` (уже в команде) |
| «0 images» / датасет пустой | указана папка с фото, а не родительская | `--train_data_dir` — папка, внутри которой лежит `2_astrakhan_style` |
| Очень медленно, десятки секунд на шаг | память вытекает в ОЗУ (особенно на Windows) | те же меры, что при нехватке памяти; в панели NVIDIA отключить «CUDA Sysmem Fallback» |
| LoRA копирует фото из датасета | переобучение | взять файл более ранней эпохи или снизить `--learning_rate` до `5e-5` |
