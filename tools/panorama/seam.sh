#!/usr/bin/env bash
# Исправление шва панорамы 2:1 той же LoRA klein 360 ERP outpaint.
#   seam.sh prep  RAW.png NAME   — сдвиг на половину ширины, зелёная полоса на месте шва → input/seam_NAME.png
#   seam.sh merge RAW.png FIXED.png OUT.png — из FIXED берётся только полоса шва (с растушёвкой), сдвиг обратно
set -euo pipefail

INPUT="${COMFY_ROOT:-$HOME/ai/ComfyUI}/input"
STRIP=160   # ширина зелёной полосы, px (при ширине 2048)
FEATHER=48  # растушёвка при вклейке, px

case "$1" in
    prep)
        raw=$2 name=$3
        read -r w h < <(magick identify -format '%w %h\n' "$raw")
        x0=$(( w / 2 - STRIP / 2 )); x1=$(( w / 2 + STRIP / 2 - 1 ))
        magick "$raw" -roll +$(( w / 2 ))+0 -fill '#00ff00' -draw "rectangle $x0,0 $x1,$h" "$INPUT/seam_$name.png"
        echo "$INPUT/seam_$name.png"
        ;;
    merge)
        raw=$2 fixed=$3 out=$4
        read -r w h < <(magick identify -format '%w %h\n' "$raw")
        half=$(( w / 2 )); band=$(( STRIP + 2 * FEATHER ))
        # Маска: белая полоса шириной STRIP+2*FEATHER по центру, края размыты.
        magick -size "${w}x${h}" xc:black -fill white \
            -draw "rectangle $(( half - band / 2 )),0 $(( half + band / 2 )),$h" \
            -blur "0x$(( FEATHER / 2 ))" /tmp/seam_mask_$$.png
        magick \( "$raw" -roll +${half}+0 \) "$fixed" /tmp/seam_mask_$$.png -composite \
            -roll -${half}+0 "$out"
        rm -f /tmp/seam_mask_$$.png
        echo "$out"
        ;;
    *) echo "usage: $0 prep RAW NAME | merge RAW FIXED OUT"; exit 1 ;;
esac
