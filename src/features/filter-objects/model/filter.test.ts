import { describe, expect, it } from 'vitest';
import { OBJECT_FIXTURES as HERITAGE_OBJECTS } from '@/entities/heritage-object/testing';
import { EMPTY_FILTER, filterObjects } from './filter';

const run = (patch: Partial<typeof EMPTY_FILTER>) => filterObjects(HERITAGE_OBJECTS, { ...EMPTY_FILTER, ...patch });

describe('filterObjects', () => {
  it('пустой фильтр возвращает все объекты', () => {
    expect(run({})).toHaveLength(HERITAGE_OBJECTS.length);
  });

  it('ищет без учёта регистра по названию, адресу и стилю', () => {
    expect(run({ query: 'КОНОНОВ' }).map((o) => o.slug)).toContain('dom-kononova');
    expect(run({ query: 'никольская' }).every((o) => o.address.toLowerCase().includes('никольская') || o.title.toLowerCase().includes('никольская') || o.architecturalStyle.toLowerCase().includes('никольская'))).toBe(true);
  });

  it('пробелы вокруг запроса не мешают', () => {
    expect(run({ query: '  кононов  ' })).toEqual(run({ query: 'кононов' }));
  });

  it('категория и статус сочетаются через «и»', () => {
    const res = run({ category: 'modern', status: 'federal' });
    expect(res.length).toBeGreaterThan(0);
    expect(res.every((o) => o.category === 'modern' && o.protectionStatus === 'federal')).toBe(true);
  });

  it('«только с 360°» оставляет объекты с панорамой', () => {
    expect(run({ only360: true }).map((o) => o.slug)).toEqual(['dom-kononova']);
  });

  it('нет совпадений — пустой список', () => {
    expect(run({ query: 'такого объекта нет' })).toEqual([]);
  });
});
