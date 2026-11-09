'use client';

import { usePathname, useRouter, useSearchParams } from 'next/navigation';

// Вкладка страницы в адресе (?tab=orders): ссылку на нужную вкладку можно открыть и переслать.
export function useTabParam<T extends string>(tabs: readonly T[], fallback: T): [T, (t: T) => void] {
  const params = useSearchParams();
  const router = useRouter();
  const pathname = usePathname() ?? '';
  const raw = params?.get('tab') as T | null;
  const tab = raw && tabs.includes(raw) ? raw : fallback;
  const setTab = (t: T) => {
    const next = new URLSearchParams(params?.toString());
    next.set('tab', t);
    router.replace(`${pathname}?${next}`, { scroll: false });
  };
  return [tab, setTab];
}
