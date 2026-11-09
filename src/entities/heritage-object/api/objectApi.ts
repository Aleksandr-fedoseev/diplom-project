import { api } from '@/shared/api';
import { fromDto } from '../lib/fromDto';
import type { HeritageObject } from '../model/types';
import type { CategoryDto, ObjectCardDto, ObjectDto, ObjectInput } from './dto';

export const objectKeys = {
  all: ['objects'] as const,
  list: (all: boolean) => ['objects', 'list', { all }] as const,
  card: (slug: string) => ['objects', 'card', slug] as const,
  categories: ['objects', 'categories'] as const,
};

export const objectApi = {
  // all=true — с неопубликованными (сервер выполнит только для модератора).
  list: async (all = false): Promise<HeritageObject[]> =>
    (await api<ObjectDto[]>('/objects', { query: { all: all || undefined } })).map((o) => fromDto(o)),
  card: async (slug: string): Promise<HeritageObject> => {
    const card = await api<ObjectCardDto>(`/objects/${encodeURIComponent(slug)}`);
    return fromDto(card.object, card.media);
  },
  categories: () => api<CategoryDto[]>('/categories'),
  createCategory: (slug: string, name: string) => api<CategoryDto>('/categories', { method: 'POST', json: { slug, name } }),
  updateCategory: (id: number, slug: string, name: string) =>
    api<CategoryDto>(`/categories/${id}`, { method: 'PUT', json: { slug, name } }),
  removeCategory: (id: number) => api<void>(`/categories/${id}`, { method: 'DELETE' }),
  create: (input: ObjectInput) => api<ObjectDto>('/objects', { method: 'POST', json: input }),
  update: (id: number, input: ObjectInput) => api<ObjectDto>(`/objects/${id}`, { method: 'PUT', json: input }),
  remove: (id: number) => api<void>(`/objects/${id}`, { method: 'DELETE' }),
  removeMedia: (fileId: number) => api<void>(`/media/${fileId}`, { method: 'DELETE' }),
  uploadMedia: (id: number, form: FormData) => api<unknown>(`/objects/${id}/media`, { method: 'POST', form }),
};
