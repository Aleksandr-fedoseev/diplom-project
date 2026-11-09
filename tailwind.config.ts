import type { Config } from 'tailwindcss';
import colors from 'tailwindcss/colors';
import plugin from 'tailwindcss/plugin';

// Тёмная тема без «dark:» у каждого класса: нейтральная и акцентные палитры заданы через
// CSS-переменные, и в тёмной теме шкала переворачивается (50 ↔ 950, 100 ↔ 900, …).
// Белая карточка с тёмным текстом становится тёмной со светлым — на всех страницах сразу.
const SHADES = [50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950] as const;
const FAMILIES = ['slate', 'blue', 'emerald', 'rose', 'amber'] as const;
type Family = (typeof FAMILIES)[number];

const rgb = (hex: string) => {
  const n = parseInt(hex.slice(1), 16);
  return `${(n >> 16) & 255} ${(n >> 8) & 255} ${n & 255}`;
};
const palette = (f: Family) => colors[f] as Record<number, string>;
const varColor = (name: string) => `rgb(var(--${name}) / <alpha-value>)`;

// Тёмная тема берёт нейтральные цвета из серой палитры neutral: у slate есть синий оттенок,
// и перевёрнутая шкала давала «синюю», а не тёмную тему.
const DARK_NEUTRAL = colors.neutral as Record<number, string>;

function themeVars(dark: boolean) {
  const vars: Record<string, string> = {};
  for (const f of FAMILIES) {
    const source = dark && f === 'slate' ? DARK_NEUTRAL : palette(f);
    SHADES.forEach((s, i) => {
      vars[`--${f}-${s}`] = rgb(source[dark ? SHADES[SHADES.length - 1 - i] : s]);
    });
    // «Чернила» — подвал и блоки с фото: в обеих темах тёмные, в тёмной — тоже без синевы.
    if (f === 'slate') {
      SHADES.forEach((s) => {
        vars[`--ink-${s}`] = rgb((dark ? DARK_NEUTRAL : palette('slate'))[s]);
      });
    }
  }
  // Свой оттенок между 800 и 900 и «белый» — фон карточек и текст на тёмных кнопках.
  vars['--slate-850'] = dark ? rgb(DARK_NEUTRAL[200]) : '21 31 50';
  vars['--white'] = dark ? rgb(DARK_NEUTRAL[900]) : '255 255 255';
  return vars;
}

const config: Config = {
  darkMode: 'class',
  content: ['./app/**/*.{js,ts,jsx,tsx,mdx}', './src/**/*.{js,ts,jsx,tsx,mdx}'],
  theme: {
    extend: {
      colors: {
        white: varColor('white'),
        // Цвета, которые не меняются с темой: подвал и текст поверх фотографий всегда тёмные/белые.
        snow: '#ffffff',
        ink: Object.fromEntries(SHADES.map((s) => [s, varColor(`ink-${s}`)])),
        slate: Object.fromEntries([...SHADES, 850].map((s) => [s, varColor(`slate-${s}`)])),
        ...Object.fromEntries(
          FAMILIES.filter((f) => f !== 'slate').map((f) => [f, Object.fromEntries(SHADES.map((s) => [s, varColor(`${f}-${s}`)]))]),
        ),
      },
      fontFamily: {
        sans: ['var(--font-inter)', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
    },
  },
  plugins: [
    plugin(({ addBase }) => {
      addBase({ ':root': themeVars(false), '.dark': { ...themeVars(true), colorScheme: 'dark' } });
    }),
  ],
};
export default config;
