import { STORIES } from '../model/data';
import type { Story } from '../model/types';

// Статьи, в которых упоминается объект (связь карты и статей, ТЗ 7.1 / 7.3).
export function storiesAboutObject(objectSlug: string): Story[] {
  return STORIES.filter((s) => s.relatedObjectSlugs?.includes(objectSlug));
}

export function findStoryBySlug(slug: string): Story | undefined {
  return STORIES.find((s) => s.slug === slug);
}
