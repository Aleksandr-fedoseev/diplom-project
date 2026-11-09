import { describe, expect, it } from 'vitest';
import type { User } from './types';
import { canBlock, canChangeRole } from './rules';

const u = (id: number, role: User['role']): User => ({ id, role, email: `${id}@t`, name: '', isBlocked: false, avatarUrl: '', createdAt: '' });
const admin = u(1, 'admin');
const mod = u(2, 'moderator');
const user = u(3, 'user');
const otherMod = u(4, 'moderator');

describe('правила управления пользователями', () => {
  it('модератор блокирует только обычных пользователей', () => {
    expect(canBlock(mod, user)).toBe(true);
    expect(canBlock(mod, otherMod)).toBe(false);
    expect(canBlock(mod, admin)).toBe(false);
  });

  it('администратор блокирует кого угодно, кроме себя', () => {
    expect(canBlock(admin, mod)).toBe(true);
    expect(canBlock(admin, admin)).toBe(false);
  });

  it('пользователь никого не блокирует', () => {
    expect(canBlock(user, u(5, 'user'))).toBe(false);
  });

  it('роль меняет только администратор и не себе', () => {
    expect(canChangeRole(admin, user)).toBe(true);
    expect(canChangeRole(admin, admin)).toBe(false);
    expect(canChangeRole(mod, user)).toBe(false);
  });
});
