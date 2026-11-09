import { api } from '@/shared/api';
import type { User } from '@/entities/user/@x/session';

export const sessionKeys = { me: ['session', 'me'] as const };

export const sessionApi = {
  login: (email: string, password: string) =>
    api<{ token: string; user: User }>('/auth/login', { method: 'POST', json: { email, password } }),
  register: (name: string, email: string, password: string) =>
    api<User>('/auth/register', { method: 'POST', json: { name, email, password } }),
  me: () => api<User>('/auth/me'),
  updateName: (name: string) => api<User>('/auth/me', { method: 'PUT', json: { name } }),
  changePassword: (currentPassword: string, newPassword: string) =>
    api<void>('/auth/me/password', { method: 'PUT', json: { currentPassword, newPassword } }),
  setAvatar: (file: File) => {
    const form = new FormData();
    form.set('file', file);
    return api<User>('/auth/me/avatar', { method: 'PUT', form });
  },
  removeAvatar: () => api<User>('/auth/me/avatar', { method: 'DELETE' }),
};
