import { describe, expect, it } from 'vitest';
import { formatDistance, formatDuration } from './format';

describe('формат маршрута', () => {
  it('расстояние: метры до километра, дальше — километры с запятой', () => {
    expect(formatDistance(347)).toBe('350 м');
    expect(formatDistance(1234)).toBe('1,2 км');
  });

  it('время: минуты и часы, не меньше минуты', () => {
    expect(formatDuration(20)).toBe('1 мин');
    expect(formatDuration(45 * 60)).toBe('45 мин');
    expect(formatDuration(60 * 60)).toBe('1 ч');
    expect(formatDuration(95 * 60)).toBe('1 ч 35 мин');
  });
});
