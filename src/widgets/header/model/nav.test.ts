import { describe, expect, it } from 'vitest';
import { accountLinks, isActiveLink, staffLinks } from './nav';

describe('isActiveLink', () => {
  it('главная активна только на «/»', () => {
    expect(isActiveLink('/', '/')).toBe(true);
    expect(isActiveLink('/', '/stories')).toBe(false);
  });

  it('раздел активен на вложенных страницах', () => {
    expect(isActiveLink('/stories', '/stories/nikolskaya')).toBe(true);
    expect(isActiveLink('/projects', '/projects/wooden-museum')).toBe(true);
  });

  it('не путает разделы с общим началом', () => {
    expect(isActiveLink('/shop', '/shopping')).toBe(false);
  });

  it('без адреса ничего не активно', () => {
    expect(isActiveLink('/shop', null)).toBe(false);
  });
});

describe('staffLinks', () => {
  const hrefs = (role?: Parameters<typeof staffLinks>[0]) => staffLinks(role).map((l) => l.href);

  it('гостю и пользователю рабочих разделов нет', () => {
    expect(hrefs(undefined)).toEqual([]);
    expect(hrefs('user')).toEqual([]);
  });

  it('модератору — модерация, администратору — ещё и админ-панель', () => {
    expect(hrefs('moderator')).toEqual(['/moderation']);
    expect(hrefs('admin')).toEqual(['/moderation', '/admin']);
  });
});

describe('accountLinks', () => {
  const hrefs = (role?: Parameters<typeof accountLinks>[0]) => accountLinks(role).map((l) => l.href);

  it('пользователю — профиль, заявки и заказы', () => {
    expect(hrefs('user')).toEqual(['/account', '/account?tab=submissions', '/account?tab=orders']);
  });

  it('сотрудникам — только профиль', () => {
    expect(hrefs('moderator')).toEqual(['/account']);
    expect(hrefs('admin')).toEqual(['/account']);
  });
});
