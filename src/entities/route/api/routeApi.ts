import { api } from '@/shared/api';
import type { RouteItem } from '../model/types';

export const routeKeys = {
  thematic: ['routes', 'thematic'] as const,
  mine: ['routes', 'mine'] as const,
  one: (slug: string) => ['routes', 'one', slug] as const,
};

export const routeApi = {
  thematic: () => api<RouteItem[]>('/routes'),
  get: (slug: string) => api<RouteItem>(`/routes/${encodeURIComponent(slug)}`),
  // Построить линию по улицам, ничего не сохраняя (доступно и гостям).
  build: (objects: string[]) => api<RouteItem>('/routes/build', { method: 'POST', json: { objects } }),
  mine: () => api<RouteItem[]>('/me/routes'),
  saveMine: (title: string, description: string, objects: string[]) =>
    api<RouteItem>('/me/routes', { method: 'POST', json: { title, description, objects } }),
  createThematic: (input: { slug: string; title: string; description: string; objects: string[] }) =>
    api<RouteItem>('/routes', { method: 'POST', json: input }),
  remove: (id: number) => api<void>(`/routes/${id}`, { method: 'DELETE' }),
};
