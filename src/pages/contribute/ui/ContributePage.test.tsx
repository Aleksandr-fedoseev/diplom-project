import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { setToken } from '@/shared/api';
import { mockApi, renderWithClient, signIn } from '@/shared/testing';
import ContributePage from './ContributePage';

vi.mock('next/navigation', () => ({ useSearchParams: () => new URLSearchParams() }));

afterEach(() => {
  setToken(null);
  vi.unstubAllGlobals();
});

describe('ContributePage', () => {
  it('гостю предлагает войти', async () => {
    mockApi({});
    renderWithClient(<ContributePage />);
    expect(await screen.findByText(/Войдите, чтобы предложить материал/)).toBeInTheDocument();
  });

  it('не пускает на следующий шаг с пустыми обязательными полями', async () => {
    mockApi(signIn({ role: 'user' }));
    const user = userEvent.setup();
    renderWithClient(<ContributePage />);
    await user.click(await screen.findByRole('button', { name: /Далее: История/ }));
    expect(screen.getByLabelText(/Название объекта/)).toBeInTheDocument();
  });

  it('проходит все шаги, сохраняет введённое и отправляет заявку с фото', async () => {
    const { calls } = mockApi({
      ...signIn({ role: 'user' }),
      'POST /submissions': ({ body }) => ({ status: 201, body: { id: 42, ...(body as object) } }),
      'POST /submissions/42/media': { status: 201, body: { id: 1 } },
    });
    const user = userEvent.setup();
    renderWithClient(<ContributePage />);

    await user.type(await screen.findByLabelText(/Название объекта/), 'Дом Франгулова');
    await user.type(screen.getByLabelText(/Адрес или ориентир/), 'ул. Советская, 10');
    await user.click(screen.getByRole('button', { name: /Далее: История/ }));

    await user.type(screen.getByLabelText(/Исторические сведения/), 'Построен в 1894 году.');
    await user.click(screen.getByRole('button', { name: /Назад/ }));
    expect(screen.getByLabelText(/Название объекта/)).toHaveValue('Дом Франгулова');
    await user.click(screen.getByRole('button', { name: /Далее: История/ }));
    await user.click(screen.getByRole('button', { name: /Далее: Фотографии/ }));

    await user.upload(screen.getByLabelText(/Нажмите для выбора/), new File(['x'], 'fasad.jpg', { type: 'image/jpeg' }));
    expect(screen.getByText('fasad.jpg')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /Отправить на модерацию/ }));

    expect(await screen.findByText('Материал успешно отправлен на модерацию')).toBeInTheDocument();
    const created = calls.find((c) => c.path === '/submissions');
    expect(created?.body).toMatchObject({ kind: 'new_object', title: 'Дом Франгулова', address: 'ул. Советская, 10' });
    expect(calls.some((c) => c.method === 'POST' && c.path === '/submissions/42/media')).toBe(true);
  });

  it('показывает ошибку сервера и остаётся на форме', async () => {
    mockApi({ ...signIn({ role: 'user' }), 'POST /submissions': { status: 403, body: { error: 'аккаунт заблокирован' } } });
    const user = userEvent.setup();
    renderWithClient(<ContributePage />);
    await user.type(await screen.findByLabelText(/Название объекта/), 'Дом');
    await user.type(screen.getByLabelText(/Адрес или ориентир/), 'ул. 1');
    await user.click(screen.getByRole('button', { name: /Далее: История/ }));
    await user.type(screen.getByLabelText(/Исторические сведения/), 'Текст');
    await user.click(screen.getByRole('button', { name: /Далее: Фотографии/ }));
    await user.click(screen.getByRole('button', { name: /Отправить на модерацию/ }));
    expect(await screen.findByText('аккаунт заблокирован')).toBeInTheDocument();
  });
});
