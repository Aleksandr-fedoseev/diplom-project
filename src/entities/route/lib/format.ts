export function formatDistance(m: number): string {
  return m < 1000 ? `${Math.round(m / 10) * 10} м` : `${(m / 1000).toFixed(1).replace('.', ',')} км`;
}

export function formatDuration(s: number): string {
  const min = Math.max(1, Math.round(s / 60));
  if (min < 60) return `${min} мин`;
  const h = Math.floor(min / 60);
  const rest = min % 60;
  return rest ? `${h} ч ${rest} мин` : `${h} ч`;
}
