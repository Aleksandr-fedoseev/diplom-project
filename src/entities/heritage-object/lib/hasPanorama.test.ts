import { describe, expect, it } from 'vitest';
import { OBJECT_FIXTURES } from '../model/fixtures';
import { hasPanorama } from './hasPanorama';

describe('hasPanorama', () => {
  const base = { ...OBJECT_FIXTURES[0], hasPanorama: false };

  it('нет признака и панорам — false', () => {
    expect(hasPanorama({ ...base, panoramas: undefined })).toBe(false);
    expect(hasPanorama({ ...base, panoramas: [] })).toBe(false);
  });

  it('признак из списка API — true', () => {
    expect(hasPanorama({ ...base, hasPanorama: true })).toBe(true);
  });

  it('есть панорамы в карточке — true', () => {
    expect(hasPanorama({ ...base, panoramas: [{ url: '/p.jpg', source: 'photo', caption: 'Вид' }] })).toBe(true);
  });

  it('в данных прототипа панорама только у дома Кононова', () => {
    expect(OBJECT_FIXTURES.filter(hasPanorama).map((o) => o.slug)).toEqual(['dom-kononova']);
  });
});
