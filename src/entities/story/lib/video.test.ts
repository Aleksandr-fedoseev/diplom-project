import { describe, expect, it } from 'vitest';
import { isTrustedVideoUrl } from './video';

describe('isTrustedVideoUrl', () => {
  it.each(['https://www.youtube-nocookie.com/embed/abc', 'https://rutube.ru/play/embed/1'])('разрешает %s', (url) => {
    expect(isTrustedVideoUrl(url)).toBe(true);
  });

  it.each([
    undefined,
    '',
    'не адрес',
    'http://www.youtube.com/embed/abc', // без https
    'https://evil.example/embed',
    'https://www.youtube.com.evil.example/embed',
    'javascript:alert(1)',
  ])('запрещает %j', (url) => {
    expect(isTrustedVideoUrl(url)).toBe(false);
  });
});
