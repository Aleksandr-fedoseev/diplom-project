// Помощники для тестов: страница рендерится с кэшем запросов, а HTTP-ответы бэкенда
// подменяются — тесты не зависят от запущенного сервера.
import type { ReactElement } from 'react';
import { render } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { vi } from 'vitest';
import { API_URL, setToken } from '@/shared/api';

export function renderWithClient(ui: ReactElement) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(<QueryClientProvider client={client}>{ui}</QueryClientProvider>);
}

type Reply = { status?: number; body?: unknown };
type Handler = Reply | ((req: { body: unknown; url: URL }) => Reply);

// mockApi({ 'GET /objects': { body: [...] }, 'POST /submissions': ({ body }) => ({ status: 201, body }) })
// Запрос без обработчика — ответ 404, а сам запрос виден в calls.
export function mockApi(handlers: Record<string, Handler>) {
  const calls: { method: string; path: string; body: unknown }[] = [];
  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = new URL(String(input));
    const method = init?.method ?? 'GET';
    const path = url.pathname.replace(new URL(API_URL).pathname, '');
    const body = typeof init?.body === 'string' ? JSON.parse(init.body) : init?.body;
    calls.push({ method, path, body });
    const h = handlers[`${method} ${path}`];
    const reply = typeof h === 'function' ? h({ body, url }) : (h ?? { status: 404, body: { error: 'не найдено' } });
    return new Response(reply.body === undefined ? null : JSON.stringify(reply.body), { status: reply.status ?? 200 });
  });
  vi.stubGlobal('fetch', fetchMock);
  return { calls, fetchMock };
}

export const DEMO_USER = {
  id: 2,
  email: 'user@demo.test',
  name: 'Иван',
  role: 'user' as const,
  isBlocked: false,
  avatarUrl: '',
  createdAt: '2026-11-01T00:00:00Z',
};

// Войти «как пользователь»: токен в хранилище и ответ /auth/me.
export function signIn(user: { role: 'user' | 'moderator' | 'admin' } & Partial<typeof DEMO_USER> = DEMO_USER) {
  setToken('test-token');
  return { 'GET /auth/me': { body: { ...DEMO_USER, ...user } } };
}
