import { describe, expect, it } from 'vitest';
import { isActiveLink } from './nav';

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
