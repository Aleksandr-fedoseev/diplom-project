import { HERITAGE_OBJECTS, type HeritageObject } from '@/entities/heritage-object/@x/route';
import type { RouteItem } from '../model/types';

// Объекты маршрута в порядке прохождения (а не в порядке каталога).
export function routeStops(route: RouteItem, objects: HeritageObject[] = HERITAGE_OBJECTS): HeritageObject[] {
  return route.objectIds.flatMap((id) => objects.filter((o) => o.id === id));
}
