#!/usr/bin/env python3
# Проекция обычного фото (камера-«обскура») на равнопромежуточную развёртку 360° (ERP 2:1).
#   control_<name>.png  — фото на зелёном холсте 2048x1024, вход для LoRA klein 360 ERP outpaint
#   overlay_<name>.png  — то же фото на прозрачном холсте 4096x2048, чтобы вернуть его поверх результата
# Использование: project_photo.py ФОТО ИМЯ HFOV_ГРАДУСЫ [НАКЛОН_ГРАДУСЫ] [ПАПКА]
import sys
from pathlib import Path

import numpy as np
from PIL import Image


def project(img: np.ndarray, out_w: int, hfov: float, pitch: float):
    """Возвращает (rgb, alpha) размера out_w x out_w/2; фото смотрит на долготу 0."""
    h, w = img.shape[:2]
    out_h = out_w // 2
    f = (w / 2) / np.tan(np.radians(hfov) / 2)

    u = (np.arange(out_w) + 0.5) / out_w
    v = (np.arange(out_h) + 0.5) / out_h
    lon = (u - 0.5) * 2 * np.pi
    lat = (0.5 - v) * np.pi
    lon, lat = np.meshgrid(lon, lat)
    dx = np.cos(lat) * np.sin(lon)
    dy = np.sin(lat)
    dz = np.cos(lat) * np.cos(lon)

    # Камера поднята на угол pitch: поворачиваем направления в её систему координат.
    p = np.radians(pitch)
    cy = dy * np.cos(p) - dz * np.sin(p)
    cz = dy * np.sin(p) + dz * np.cos(p)

    front = cz > 1e-6
    with np.errstate(divide="ignore", invalid="ignore"):
        px = f * dx / cz + w / 2 - 0.5
        py = -f * cy / cz + h / 2 - 0.5
    inside = front & (px >= 0) & (px <= w - 1) & (py >= 0) & (py <= h - 1)

    # Билинейная выборка.
    x0 = np.clip(np.floor(px), 0, w - 2).astype(int)
    y0 = np.clip(np.floor(py), 0, h - 2).astype(int)
    fx = np.clip(px - x0, 0, 1)[..., None]
    fy = np.clip(py - y0, 0, 1)[..., None]
    x0[~inside] = 0
    y0[~inside] = 0
    src = img.astype(np.float32)
    top = src[y0, x0] * (1 - fx) + src[y0, x0 + 1] * fx
    bot = src[y0 + 1, x0] * (1 - fx) + src[y0 + 1, x0 + 1] * fx
    rgb = top * (1 - fy) + bot * fy

    # Мягкий край маски (несколько пикселей фото), чтобы шов с дорисовкой не был резким.
    edge = np.minimum.reduce([px, w - 1 - px, py, h - 1 - py])
    alpha = np.where(inside, np.clip(edge / (0.02 * w), 0, 1), 0.0)
    return rgb, alpha


def keep_mask(img: np.ndarray) -> np.ndarray:
    """1 — пиксель фото кладётся поверх панорамы, 0 — небо, его оставляем нейросети.

    Небо ищется в каждом столбце сверху вниз: светлый, не красный и гладкий пиксель.
    Первый пиксель, который этому не отвечает, — верх здания (дерева, столба)."""
    h, w = img.shape[:2]
    f = img.astype(np.float32)
    r, g, b = f[..., 0], f[..., 1], f[..., 2]
    bright = (r + g + b) / 3
    grad = np.zeros((h, w), np.float32)
    grad[:-1] += np.abs(np.diff(bright, axis=0))
    grad[:, :-1] += np.abs(np.diff(bright, axis=1))
    # светлое (белое, серое, голубое) или явно синее, даже тёмное; кирпич и зелень не проходят
    sky = (b >= r - 8) & ((bright > 110) | (b > r + 40)) & (grad < 14)

    # Высота неба в столбце: подряд идущие пиксели неба от верхнего края (тонкие разрывы — провода).
    gap = max(3, h // 150)
    solid = np.logical_and.reduce([~sky[i:h - gap + i] for i in range(gap + 1)])
    run = np.where(solid.any(axis=0), solid.argmax(axis=0), h)

    # Столбы и одиночные ветки дают узкие провалы — сглаживаем медианой по ~2% ширины.
    k = max(3, w // 50) | 1
    pad = np.pad(run, k // 2, mode="edge")
    run = np.median(np.lib.stride_tricks.sliding_window_view(pad, k), axis=1)

    # Плавный переход вверх от линии крыш (фото тает в небе), ширина ~2% высоты.
    feather = max(4.0, 0.02 * h)
    y = np.arange(h)[:, None]
    return np.clip(1 - (run[None, :] - y) / feather, 0, 1).astype(np.float32)


def main():
    photo, name, hfov = sys.argv[1], sys.argv[2], float(sys.argv[3])
    pitch = float(sys.argv[4]) if len(sys.argv) > 4 else 0.0
    out_dir = Path(sys.argv[5] if len(sys.argv) > 5 else ".")
    img = np.asarray(Image.open(photo).convert("RGB"))

    rgb, alpha = project(img, 2048, hfov, pitch)
    green = np.array([0, 255, 0], np.float32)
    hard = (alpha > 0)[..., None]  # на холсте управления фото без растушёвки
    control = np.where(hard, rgb, green)
    Image.fromarray(control.round().astype(np.uint8)).save(out_dir / f"control_{name}.png")

    # На overlay небо с фото не кладём: его цвет почти никогда не совпадает с дорисованным.
    keep = keep_mask(img)
    rgbk, alpha = project(np.dstack([img, keep * 255]), 4096, hfov, pitch)
    rgb, alpha = rgbk[..., :3], alpha * rgbk[..., 3] / 255
    rgba = np.dstack([rgb, alpha * 255]).round().astype(np.uint8)
    Image.fromarray(rgba, "RGBA").save(out_dir / f"overlay_{name}.png")
    print(name, f"фото занимает {hard.mean() * 100:.1f}% панорамы")


if __name__ == "__main__":
    main()
