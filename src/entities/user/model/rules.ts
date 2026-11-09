import type { User } from './types';

type Actor = Pick<User, 'id' | 'role'>;

// Повторяют правила бэкенда (domain/auth: CheckBlock, CheckRoleChange), чтобы не показывать
// кнопки, которые сервер всё равно отклонит.

// Блокировать: не себя; модератор — только обычных пользователей, администратор — кого угодно.
export const canBlock = (actor: Actor, target: User) =>
  actor.id !== target.id && (actor.role === 'admin' || (actor.role === 'moderator' && target.role === 'user'));

// Менять роль: только администратор и не себе.
export const canChangeRole = (actor: Actor, target: User) => actor.role === 'admin' && actor.id !== target.id;
