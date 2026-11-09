// Адрес API бэкенда. В .env.local фронтенда: NEXT_PUBLIC_API_URL=http://localhost:8080/api
export const API_URL = (process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api').replace(/\/$/, '');

// Бэкенд отдаёт адреса файлов от корня сервера (/api/media/5/file) — добавляем адрес сервера.
const API_ORIGIN = API_URL.replace(/\/api$/, '');

export function mediaUrl(path: string): string {
  if (!path) return '';
  return /^https?:\/\//.test(path) ? path : `${API_ORIGIN}${path}`;
}
