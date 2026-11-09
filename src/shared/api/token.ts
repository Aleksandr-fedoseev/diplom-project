// Токен входа. Хранится в localStorage, чтобы вход переживал перезагрузку страницы.
// Подписчики узнают о входе и выходе — так шапка и страницы обновляются сразу.
const KEY = 'heritage.token';
const listeners = new Set<() => void>();

export function getToken(): string | null {
  if (typeof window === 'undefined') return null;
  try {
    return localStorage.getItem(KEY);
  } catch {
    return null;
  }
}

export function setToken(token: string | null) {
  try {
    if (token) localStorage.setItem(KEY, token);
    else localStorage.removeItem(KEY);
  } catch {
    // приватный режим браузера: вход продержится до перезагрузки страницы
  }
  listeners.forEach((l) => l());
}

export function subscribeToken(listener: () => void): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}
