'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Search } from 'lucide-react';

// Поиск по памятникам: открывает каталог с запросом (/objects?q=…).
export function SiteSearch({ onDone }: { onDone: () => void }) {
  const [query, setQuery] = useState('');
  const router = useRouter();

  return (
    <form
      role="search"
      onSubmit={(e) => {
        e.preventDefault();
        const q = query.trim();
        if (!q) return;
        router.push(`/objects?q=${encodeURIComponent(q)}`);
        onDone();
      }}
      className="border-t border-slate-200 bg-slate-50 px-4 py-3 animate-in fade-in slide-in-from-top-2 duration-150"
    >
      <div className="max-w-3xl mx-auto relative">
        <Search className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
        <input
          type="search"
          autoFocus
          aria-label="Поиск по памятникам"
          placeholder="Поиск по памятникам: название, улица, стиль…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => e.key === 'Escape' && onDone()}
          className="w-full bg-white border border-slate-200 pl-10 pr-24 py-2.5 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950 [&::-webkit-search-cancel-button]:hidden"
        />
        <button
          type="submit"
          className="absolute right-1.5 top-1/2 -translate-y-1/2 px-3 py-1.5 bg-slate-950 hover:bg-slate-800 text-white rounded-md text-xs font-bold"
        >
          Найти
        </button>
      </div>
    </form>
  );
}
