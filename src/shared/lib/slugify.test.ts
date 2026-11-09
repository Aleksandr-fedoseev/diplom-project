import { describe, expect, it } from 'vitest';
import { slugify } from './slugify';

describe('slugify', () => {
  it('переводит кириллицу в латиницу через дефис', () => {
    expect(slugify('Дом купца Тетюшинова')).toBe('dom-kuptsa-tetyushinova');
    expect(slugify('Усадьба А.И. Губина')).toBe('usadba-a-i-gubina');
  });

  it('без лишних дефисов по краям и подряд', () => {
    expect(slugify('  «Биржа» — 1906!  ')).toBe('birzha-1906');
  });
});
