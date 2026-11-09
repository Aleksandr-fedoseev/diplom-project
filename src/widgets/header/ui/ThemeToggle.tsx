'use client';

import { Moon, Sun } from 'lucide-react';
import { useTheme } from '@/shared/theme';

const iconButton = 'p-2 text-slate-600 hover:text-slate-950 hover:bg-slate-100 rounded-lg transition-colors';

// Переключатель светлой и тёмной темы. До первого нажатия тема — как в системе.
export function ThemeToggle() {
  const { dark, setChoice } = useTheme();
  return (
    <button
      type="button"
      onClick={() => setChoice(dark ? 'light' : 'dark')}
      className={iconButton}
      aria-label={dark ? 'Светлая тема' : 'Тёмная тема'}
      title={dark ? 'Светлая тема' : 'Тёмная тема'}
    >
      {/* Иконку выбирает CSS по классу dark — она верна с первой отрисовки. */}
      <Sun className="w-5 h-5 hidden dark:block" />
      <Moon className="w-5 h-5 dark:hidden" />
    </button>
  );
}
