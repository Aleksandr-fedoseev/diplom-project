import { describe, expect, it } from 'vitest';
import { ROUTES } from '../model/data';
import { routeStops } from './stops';

describe('routeStops', () => {
  it('возвращает остановки в порядке маршрута', () => {
    const route = { ...ROUTES[0], objectIds: ['dom-tavrizova', 'birzha'] };
    expect(routeStops(route).map((o) => o.id)).toEqual(['dom-tavrizova', 'birzha']);
  });

  it('пропускает неизвестные объекты', () => {
    expect(routeStops({ ...ROUTES[0], objectIds: ['нет-такого', 'birzha'] }).map((o) => o.id)).toEqual(['birzha']);
  });

  it('у всех маршрутов все остановки существуют', () => {
    for (const r of ROUTES) expect(routeStops(r)).toHaveLength(r.objectIds.length);
  });
});
