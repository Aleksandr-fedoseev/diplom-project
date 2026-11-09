// Маршрут в формате API (internal/transport/http/route). Расстояние и время считает OSRM.
export interface RouteStop {
  position: number;
  objectId: number;
  slug: string;
  title: string;
  lat: number;
  lng: number;
}

export interface RouteItem {
  id: number;
  slug: string;
  title: string;
  description: string;
  ownerId?: number; // нет — тематический маршрут администратора
  distanceM: number;
  durationS: number;
  geometry?: { type: 'LineString'; coordinates: [number, number][] }; // только в карточке маршрута
  stops: RouteStop[];
  createdAt?: string;
}
