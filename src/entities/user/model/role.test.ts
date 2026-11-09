import { describe, expect, it } from 'vitest';
import { isStaff, roleAtLeast } from './role';

describe('roleAtLeast', () => {
  it('старшая роль включает права младшей', () => {
    expect(roleAtLeast('admin', 'moderator')).toBe(true);
    expect(roleAtLeast('moderator', 'moderator')).toBe(true);
    expect(roleAtLeast('moderator', 'admin')).toBe(false);
    expect(roleAtLeast('user', 'moderator')).toBe(false);
  });

  it('гость не проходит ни одну проверку', () => {
    expect(roleAtLeast(undefined, 'user')).toBe(false);
  });
});

describe('isStaff', () => {
  it('сотрудники — модератор и администратор', () => {
    expect(isStaff('moderator')).toBe(true);
    expect(isStaff('admin')).toBe(true);
    expect(isStaff('user')).toBe(false);
    expect(isStaff(undefined)).toBe(false);
  });
});
