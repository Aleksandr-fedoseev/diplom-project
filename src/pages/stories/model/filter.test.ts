import { describe, expect, it } from 'vitest';
import { STORIES } from '@/entities/story';
import { filterStories } from './filter';

describe('filterStories', () => {
  it('без условий — все статьи', () => {
    expect(filterStories(STORIES, 'all', '')).toHaveLength(STORIES.length);
  });

  it('по рубрике', () => {
    const res = filterStories(STORIES, 'wooden', '');
    expect(res.length).toBeGreaterThan(0);
    expect(res.every((s) => s.category === 'wooden')).toBe(true);
  });

  it('по тексту заголовка без учёта регистра', () => {
    const word = STORIES[0].title.split(' ')[0].toUpperCase();
    expect(filterStories(STORIES, 'all', word)).toContain(STORIES[0]);
  });

  it('пустая рубрика — пустой список', () => {
    expect(filterStories(STORIES, 'news', '')).toEqual([]);
  });
});
