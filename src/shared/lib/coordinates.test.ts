import { describe, expect, it } from 'vitest';
import { parseCoordinates } from './coordinates';

describe('parseCoordinates', () => {
  it('понимает точку и запятую-разделитель', () => {
    expect(parseCoordinates('46.3538, 48.0285')).toEqual({ lat: 46.3538, lng: 48.0285 });
    expect(parseCoordinates('46.3538 48.0285')).toEqual({ lat: 46.3538, lng: 48.0285 });
    expect(parseCoordinates('46,3538 48,0285')).toEqual({ lat: 46.3538, lng: 48.0285 });
    expect(parseCoordinates('46,3538; 48,0285')).toEqual({ lat: 46.3538, lng: 48.0285 });
  });

  it('пусто или ерунда — null', () => {
    expect(parseCoordinates('')).toBeNull();
    expect(parseCoordinates('где-то у Кремля')).toBeNull();
    expect(parseCoordinates('46.35')).toBeNull();
    expect(parseCoordinates('146.35, 48.02')).toBeNull();
  });
});
