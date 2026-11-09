// Тема сайта: светлая, тёмная или как в системе. Выбор хранится в браузере.
export type ThemeChoice = 'light' | 'dark' | 'system';

export const THEME_KEY = 'heritage.theme';

const listeners = new Set<() => void>();

export function getThemeChoice(): ThemeChoice {
  if (typeof window === 'undefined') return 'system';
  try {
    const v = localStorage.getItem(THEME_KEY);
    return v === 'light' || v === 'dark' ? v : 'system';
  } catch {
    return 'system';
  }
}

const systemDark = () => typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches;

export const isDark = (choice: ThemeChoice) => choice === 'dark' || (choice === 'system' && systemDark());

export function applyTheme(choice: ThemeChoice) {
  document.documentElement.classList.toggle('dark', isDark(choice));
}

export function setThemeChoice(choice: ThemeChoice) {
  try {
    if (choice === 'system') localStorage.removeItem(THEME_KEY);
    else localStorage.setItem(THEME_KEY, choice);
  } catch {
    // приватный режим: тема продержится до перезагрузки
  }
  applyTheme(choice);
  listeners.forEach((l) => l());
}

export function subscribeTheme(listener: () => void): () => void {
  listeners.add(listener);
  // Смена темы в системе при выборе «как в системе».
  const mq = window.matchMedia('(prefers-color-scheme: dark)');
  const onSystem = () => {
    if (getThemeChoice() === 'system') applyTheme('system');
    listener();
  };
  mq.addEventListener('change', onSystem);
  return () => {
    listeners.delete(listener);
    mq.removeEventListener('change', onSystem);
  };
}

// Выполняется в <head> до отрисовки: страница сразу открывается в нужной теме, без вспышки белого.
export const THEME_INIT_SCRIPT = `(function(){try{var t=localStorage.getItem('${THEME_KEY}');var d=t==='dark'||(t!=='light'&&matchMedia('(prefers-color-scheme: dark)').matches);if(d)document.documentElement.classList.add('dark')}catch(e){}})()`;
