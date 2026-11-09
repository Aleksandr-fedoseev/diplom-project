import type { Role } from './types';

const LEVEL: Record<Role, number> = { user: 1, moderator: 2, admin: 3 };

export const ROLE_LABELS: Record<Role, string> = {
  user: 'Пользователь',
  moderator: 'Модератор',
  admin: 'Администратор',
};

// Повторяет правило бэкенда Role.AtLeast. Интерфейс только прячет недоступное —
// права всё равно проверяет сервер.
export const roleAtLeast = (role: Role | undefined, min: Role) => !!role && LEVEL[role] >= LEVEL[min];

// Сотрудник — модератор или администратор. Работает через «Модерацию» и «Админ-панель»;
// личные заявки и заказы — с обычного аккаунта.
export const isStaff = (role: Role | undefined) => roleAtLeast(role, 'moderator');
