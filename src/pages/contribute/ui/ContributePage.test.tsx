import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import ContributePage from './ContributePage';

describe('ContributePage', () => {
  it('не пускает на следующий шаг с пустыми обязательными полями', async () => {
    const user = userEvent.setup();
    render(<ContributePage />);
    await user.click(screen.getByRole('button', { name: /Далее: История/ }));
    expect(screen.getByLabelText(/Название объекта/)).toBeInTheDocument();
  });

  it('проходит все шаги, сохраняет введённое при возврате и отправляет заявку', async () => {
    const user = userEvent.setup();
    render(<ContributePage />);

    await user.type(screen.getByLabelText(/Название объекта/), 'Дом Франгулова');
    await user.type(screen.getByLabelText(/Адрес или ориентир/), 'ул. Советская, 10');
    await user.click(screen.getByRole('button', { name: /Далее: История/ }));

    await user.type(screen.getByLabelText(/Исторические сведения/), 'Построен в 1894 году.');
    await user.click(screen.getByRole('button', { name: /Назад/ }));
    expect(screen.getByLabelText(/Название объекта/)).toHaveValue('Дом Франгулова');
    await user.click(screen.getByRole('button', { name: /Далее: История/ }));
    expect(screen.getByLabelText(/Исторические сведения/)).toHaveValue('Построен в 1894 году.');
    await user.click(screen.getByRole('button', { name: /Далее: Загрузка медиа/ }));

    await user.type(screen.getByLabelText(/Ваше имя/), 'Иван');
    await user.type(screen.getByLabelText(/Контактный Email/), 'ivan@example.ru');
    await user.click(screen.getByRole('button', { name: /Отправить на модерацию/ }));

    expect(screen.getByText('Материал успешно отправлен на модерацию')).toBeInTheDocument();
  });
});
