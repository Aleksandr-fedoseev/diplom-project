import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { ContributionForm } from './ContributionForm';

describe('ContributionForm', () => {
  it('после отправки благодарит и называет контакт', async () => {
    const user = userEvent.setup();
    render(<ContributionForm kind="archive" />);

    await user.type(screen.getByLabelText('Ваше имя*'), 'Елена');
    await user.type(screen.getByLabelText('Контактный телефон или Email*'), 'elena@example.ru');
    await user.type(screen.getByLabelText(/Описание материалов/), 'Фото набережной, 1910-е');
    await user.click(screen.getByRole('button', { name: 'Отправить материалы в архив' }));

    expect(screen.getByText('Спасибо за ваш вклад в историю Астрахани!')).toBeInTheDocument();
    expect(screen.getByText(/elena@example\.ru/)).toBeInTheDocument();
  });

  it('не отправляет форму с пустыми обязательными полями', async () => {
    const user = userEvent.setup();
    render(<ContributionForm kind="story" />);

    await user.click(screen.getByRole('button', { name: 'Отправить историю в редакцию' }));

    expect(screen.queryByText('Спасибо за ваш вклад в историю Астрахани!')).not.toBeInTheDocument();
  });

  it('поле адреса есть только у видов, привязанных к месту', () => {
    const { unmount } = render(<ContributionForm kind="archive" />);
    expect(screen.queryByLabelText(/Адрес|Улица/)).not.toBeInTheDocument();
    unmount();

    render(<ContributionForm kind="wooden" />);
    expect(screen.getByLabelText('Адрес происхождения дома в Астрахани*')).toBeInTheDocument();
  });
});
