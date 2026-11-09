// Кросс-импорт по правилам FSD: сессия — это текущий пользователь и его роль.
export type { Role, User } from '../model/types';
export { isStaff, ROLE_LABELS, roleAtLeast } from '../model/role';
