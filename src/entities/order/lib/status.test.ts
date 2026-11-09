import { describe, expect, it } from 'vitest';
import { nextOrderStatuses } from './status';

describe('nextOrderStatuses', () => {
  it('повторяет правила бэкенда', () => {
    expect(nextOrderStatuses('pending')).toEqual(['cancelled']);
    expect(nextOrderStatuses('paid')).toEqual(['shipped', 'cancelled']);
    expect(nextOrderStatuses('shipped')).toEqual(['completed']);
  });

  it('из конечных статусов переходов нет', () => {
    expect(nextOrderStatuses('completed')).toEqual([]);
    expect(nextOrderStatuses('cancelled')).toEqual([]);
  });
});
