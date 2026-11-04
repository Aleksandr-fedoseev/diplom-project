// 1950 → «1 950 ₽» (неразрывный пробел между разрядами, как принято в русской типографике).
export function formatRub(amount: number): string {
  return `${amount.toLocaleString('ru-RU')} ₽`;
}
