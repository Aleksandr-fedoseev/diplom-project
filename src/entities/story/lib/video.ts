// Встраивать в страницу можно только видео с известных сервисов: адрес приходит из данных
// (а потом из API), и произвольный сайт во фрейме — риск для посетителя.
const TRUSTED_HOSTS = ['www.youtube-nocookie.com', 'www.youtube.com', 'rutube.ru', 'vk.com', 'vkvideo.ru'];

export function isTrustedVideoUrl(url: string | undefined): url is string {
  if (!url) return false;
  try {
    const u = new URL(url);
    return u.protocol === 'https:' && TRUSTED_HOSTS.includes(u.hostname);
  } catch {
    return false;
  }
}
