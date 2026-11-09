// «46.3538, 48.0285» (широта, долгота) → { lat, lng }. Подходит и запятая-разделитель дробной
// части через пробел: «46,3538 48,0285». Только Астраханская область не проверяется — это
// забота сервера и модератора; здесь — только что числа правдоподобны.
export function parseCoordinates(text: string): { lat: number; lng: number } | null {
  const t = text.trim();
  if (!t) return null;
  const parts = t.includes(';') || /\d,\d+\s+\d/.test(t) ? t.split(/[;\s]+/) : t.split(/[,\s]+/);
  const nums = parts.filter(Boolean).map((p) => Number(p.replace(',', '.')));
  if (nums.length !== 2 || nums.some((n) => !Number.isFinite(n))) return null;
  const [lat, lng] = nums;
  if (Math.abs(lat) > 90 || Math.abs(lng) > 180) return null;
  return { lat, lng };
}
