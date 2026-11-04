import type { Story } from '@/entities/story';

// Рубрика и поиск по заголовку и лиду, без учёта регистра.
export function filterStories(stories: Story[], category: string, query: string): Story[] {
  const q = query.trim().toLowerCase();
  return stories.filter(
    (s) =>
      (category === 'all' || s.category === category) &&
      (q === '' || s.title.toLowerCase().includes(q) || s.lead.toLowerCase().includes(q)),
  );
}
