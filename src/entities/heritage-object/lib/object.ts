import { HERITAGE_OBJECTS } from '../model/data';
import type { HeritageObject } from '../model/types';

export function findObjectBySlug(slug: string): HeritageObject | undefined {
  return HERITAGE_OBJECTS.find((o) => o.slug === slug);
}

// Ссылка на карту с открытым объектом.
export function objectMapHref(object: Pick<HeritageObject, 'slug'>): string {
  return `/map?object=${encodeURIComponent(object.slug)}`;
}

export function formatCoordinates({ lat, lng }: HeritageObject['coordinates']): string {
  return `${lat.toFixed(4)}° N, ${lng.toFixed(4)}° E`;
}
