import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { TOURS } from '@/entities/tour';
import ToursPage from './ToursPage';

describe('ToursPage', () => {
  it('«Записаться в группу» выбирает эту экскурсию в форме записи', async () => {
    const user = userEvent.setup();
    render(<ToursPage />);
    const groupTours = TOURS.filter((t) => !t.isOnlineRoute);
    const last = groupTours[groupTours.length - 1];

    await user.click(screen.getAllByRole('link', { name: /Записаться в группу/ }).at(-1)!);

    expect(screen.getByLabelText('Выбранная экскурсия*')).toHaveValue(last.id);
  });

  it('онлайн-маршрут не предлагается в форме записи', () => {
    render(<ToursPage />);
    const options = Array.from((screen.getByLabelText('Выбранная экскурсия*') as HTMLSelectElement).options).map((o) => o.value);
    expect(options).toEqual(TOURS.filter((t) => !t.isOnlineRoute).map((t) => t.id));
  });
});
