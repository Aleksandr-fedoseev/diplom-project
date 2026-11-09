// Роли совпадают с бэкендом: каждая следующая включает права предыдущей.
export type Role = 'user' | 'moderator' | 'admin';

export interface User {
  id: number;
  email: string;
  name: string;
  role: Role;
  isBlocked: boolean;
  avatarUrl: string; // пусто — аватара нет
  createdAt: string;
}
