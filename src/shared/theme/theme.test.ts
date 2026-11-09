import { afterEach, describe, expect, it } from 'vitest';
import { getThemeChoice, setThemeChoice } from './theme';

afterEach(() => setThemeChoice('system'));

describe('тема', () => {
  it('по умолчанию — как в системе', () => {
    expect(getThemeChoice()).toBe('system');
  });

  it('выбор сохраняется и включает класс dark', () => {
    setThemeChoice('dark');
    expect(getThemeChoice()).toBe('dark');
    expect(document.documentElement.classList.contains('dark')).toBe(true);
    setThemeChoice('light');
    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });
});
