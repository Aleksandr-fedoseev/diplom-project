import { describe, expect, it } from 'vitest';
import { HERITAGE_OBJECTS } from '../model/data';
import { hasPanorama } from './hasPanorama';

describe('hasPanorama', () => {
  const base = HERITAGE_OBJECTS[0];

  it('нет списка панорам — нет панорамы', () => {
    expect(hasPanorama({ ...base, panoramas: undefined })).toBe(false);
  });

  it('пустой список — нет панорамы', () => {
    expect(hasPanorama({ ...base, panoramas: [] })).toBe(false);
  });

  it('есть хотя бы одна панорама', () => {
    expect(hasPanorama({ ...base, panoramas: [{ url: '/p.jpg', source: 'photo', caption: 'Вид' }] })).toBe(true);
  });

  it('значок 360° сейчас только у дома Кононова', () => {
    expect(HERITAGE_OBJECTS.filter(hasPanorama).map((o) => o.slug)).toEqual(['dom-kononova']);
  });
});
