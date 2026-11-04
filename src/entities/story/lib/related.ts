import { STORIES } from '../model/data';
import type { Story } from '../model/types';

// Статьи, в которых упоминается объект (связь карты и статей, ТЗ 7.1 / 7.3).
export function storiesAboutObject(objectSlug: string): Story[] {
  return STORIES.filter((s) => s.relatedObjectSlugs?.includes(objectSlug));
}

export function findStoryBySlug(slug: string): Story | undefined {
  return STORIES.find((s) => s.slug === slug);
}

// Статьи из блока «Связанные статьи» в заданном автором порядке.
export function relatedStories(story: Story): Story[] {
  return (story.relatedStorySlugs ?? []).flatMap((slug) => STORIES.filter((s) => s.slug === slug));
}
