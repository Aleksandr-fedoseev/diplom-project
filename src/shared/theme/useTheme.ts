'use client';

import { useSyncExternalStore } from 'react';
import { getThemeChoice, isDark, setThemeChoice, subscribeTheme, type ThemeChoice } from './theme';

export function useTheme() {
  const choice = useSyncExternalStore(subscribeTheme, getThemeChoice, (): ThemeChoice => 'system');
  return { choice, dark: typeof window !== 'undefined' && isDark(choice), setChoice: setThemeChoice };
}
