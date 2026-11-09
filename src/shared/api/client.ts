import { API_URL } from './config';
import { getToken, setToken } from './token';

// Ошибка API: код HTTP и текст из ответа бэкенда ({"error": "..."}).
export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  json?: unknown;
  form?: FormData;
  query?: Record<string, string | number | boolean | undefined>;
  signal?: AbortSignal;
}

const FALLBACK: Record<number, string> = {
  401: 'Нужно войти в аккаунт',
  403: 'Недостаточно прав',
  404: 'Не найдено',
  429: 'Слишком много попыток, подождите немного',
};

export async function api<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const url = new URL(API_URL + path);
  for (const [k, v] of Object.entries(opts.query ?? {})) {
    if (v !== undefined && v !== '') url.searchParams.set(k, String(v));
  }

  const headers: Record<string, string> = {};
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  let body: BodyInit | undefined;
  if (opts.json !== undefined) {
    headers['Content-Type'] = 'application/json';
    body = JSON.stringify(opts.json);
  } else if (opts.form) {
    body = opts.form; // границу multipart браузер ставит сам
  }

  let res: Response;
  try {
    res = await fetch(url, { method: opts.method ?? 'GET', headers, body, signal: opts.signal });
  } catch (e) {
    if ((e as Error).name === 'AbortError') throw e;
    throw new ApiError(0, 'Сервер недоступен. Проверьте, что бэкенд запущен.');
  }

  if (res.status === 401 && token) setToken(null); // токен истёк или пользователь заблокирован

  if (!res.ok) {
    const data = await res.json().catch(() => null);
    throw new ApiError(res.status, data?.error || FALLBACK[res.status] || `Ошибка сервера (${res.status})`);
  }
  if (res.status === 204) return undefined as T;
  const text = await res.text();
  return (text ? JSON.parse(text) : undefined) as T;
}

export const errorMessage = (e: unknown) =>
  e instanceof ApiError ? e.message : e instanceof Error ? e.message : 'Неизвестная ошибка';
