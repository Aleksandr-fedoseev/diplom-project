import { api } from '@/shared/api';
import type { Role, User } from '../model/types';

export const userKeys = { list: ['users'] as const };

// Управление пользователями: список и роли — администратор, блокировка — модератор и выше.
export const userApi = {
  list: () => api<User[]>('/admin/users', { query: { limit: 200 } }),
  setRole: (id: number, role: Role) => api<User>(`/admin/users/${id}/role`, { method: 'PUT', json: { role } }),
  setBlocked: (id: number, blocked: boolean) =>
    api<User>(`/admin/users/${id}/blocked`, { method: 'PUT', json: { blocked } }),
};
