export const NAV_LINKS = [
  { href: '/', label: 'Главная' },
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
