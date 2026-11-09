import type { HeritageObject } from '../model/types';

// Ссылка на карту с открытым объектом.
export function objectMapHref(object: Pick<HeritageObject, 'slug'>): string {
  return `/map?object=${encodeURIComponent(object.slug)}`;
}

export function formatCoordinates({ lat, lng }: HeritageObject['coordinates']): string {
  return `${lat.toFixed(4)}° N, ${lng.toFixed(4)}° E`;
}

// «1906–1910 г. • Модерн» — без пустых частей, если год или стиль неизвестны.
export function objectSubtitle(o: Pick<HeritageObject, 'yearBuilt' | 'architecturalStyle'>): string {
  return [o.yearBuilt && `${o.yearBuilt} г.`, o.architecturalStyle].filter(Boolean).join(' • ');
}
