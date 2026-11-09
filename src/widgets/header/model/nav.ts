import { isStaff, roleAtLeast, type Role } from '@/entities/user';

export const NAV_LINKS = [
  { href: '/', label: 'Главная' },
  { href: '/map', label: 'Карта' },
  { href: '/stories', label: 'Истории' },
  { href: '/projects', label: 'Проекты' },
  { href: '/shop', label: 'Мерч' },
  { href: '/tours', label: 'Экскурсии' },
  { href: '/contacts', label: 'Где нас найти' },
];

// Раздел активен и на вложенных страницах (/stories/… → «Истории»), а главная — только на «/».
export function isActiveLink(href: string, pathname: string | null): boolean {
  if (!pathname) return false;
  return href === '/' ? pathname === '/' : pathname === href || pathname.startsWith(`${href}/`);
}

// Рабочие разделы сотрудников — отдельными вкладками в шапке. Администратор наследует
// права модератора (диаграмма прецедентов), поэтому видит обе вкладки.
export function staffLinks(role: Role | undefined): { href: string; label: string }[] {
  return [
    ...(roleAtLeast(role, 'moderator') ? [{ href: '/moderation', label: 'Модерация' }] : []),
    ...(roleAtLeast(role, 'admin') ? [{ href: '/admin', label: 'Админ-панель' }] : []),
  ];
}

// Меню пользователя: профиль, а у обычного пользователя — ещё его заявки и заказы.
// Сотрудникам они не нужны: их работа — во вкладках «Модерация» и «Админ-панель».
export function accountLinks(role: Role | undefined): { href: string; label: string }[] {
  if (!role) return [];
  return [
    { href: '/account', label: 'Профиль' },
    ...(isStaff(role)
      ? []
      : [
          { href: '/account?tab=submissions', label: 'Мои заявки' },
          { href: '/account?tab=orders', label: 'Мои заказы' },
        ]),
  ];
}
