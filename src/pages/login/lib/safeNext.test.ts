import { describe, expect, it } from 'vitest';
import { safeNext } from './safeNext';

describe('safeNext', () => {
  it('пускает адреса этого сайта', () => {
    expect(safeNext('/shop')).toBe('/shop');
    expect(safeNext('/contribute?object=birzha')).toBe('/contribute?object=birzha');
  });

  it('чужие адреса и пустое значение ведут в кабинет', () => {
    for (const bad of [null, '', 'https://evil.ru', '//evil.ru', '/\\evil.ru', 'javascript:alert(1)']) {
      expect(safeNext(bad)).toBe('/account');
    }
  });
});
