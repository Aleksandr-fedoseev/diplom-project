import { describe, expect, it } from 'vitest';
import { plural } from './plural';

const forms: [string, string, string] = ['объект', 'объекта', 'объектов'];

describe('plural', () => {
  it.each([
    [1, 'объект'], [21, 'объект'], [101, 'объект'],
    [2, 'объекта'], [4, 'объекта'], [22, 'объекта'],
    [0, 'объектов'], [5, 'объектов'], [11, 'объектов'], [12, 'объектов'], [14, 'объектов'], [111, 'объектов'],
  ])('%i → %s', (n, expected) => {
    expect(plural(n, forms)).toBe(expected);
  });
});
