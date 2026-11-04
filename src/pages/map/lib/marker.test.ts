import { describe, expect, it } from 'vitest';
import { HERITAGE_OBJECTS } from '@/entities/heritage-object';
import { escapeHtml, markerHtml } from './marker';

describe('markerHtml', () => {
  const base = HERITAGE_OBJECTS[0];

  it('экранирует название объекта', () => {
    const html = markerHtml({ ...base, title: '<img src=x onerror=alert(1)>' }, false);
    expect(html).not.toContain('<img');
    expect(html).toContain('&lt;img src=x onerror=alert(1)&gt;');
  });

  it('отмечает выбранный и утраченный объект', () => {
    expect(markerHtml({ ...base, category: 'lost' }, true)).toContain('custom-osm-pin is-active is-lost');
  });

  it('escapeHtml заменяет все опасные символы', () => {
    expect(escapeHtml(`&<>"'`)).toBe('&amp;&lt;&gt;&quot;&#39;');
  });
});
