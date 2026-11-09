// Куда вернуть после входа: только адрес этого сайта. «//evil.ru» и «https://...»
// увели бы пользователя на чужой сайт (открытое перенаправление).
export function safeNext(next: string | null | undefined): string {
  if (!next || !next.startsWith('/') || next.startsWith('//') || next.startsWith('/\\')) return '/account';
  return next;
}
