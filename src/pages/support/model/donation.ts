// Готовые суммы пожертвования, ₽.
export const DONATION_AMOUNTS = ['300', '500', '1000', '2500'] as const;

// Итоговая сумма: введённая вручную важнее выбранной кнопкой, если это положительное число.
export function donationSum(preset: string, custom: string): string {
  const value = Number(custom);
  return custom.trim() !== '' && Number.isFinite(value) && value > 0 ? String(value) : preset;
}
