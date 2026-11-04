import { BookMarked } from 'lucide-react';
import type { StorySource } from '@/entities/story';

// Источники статьи (ТЗ 5.2).
export function StorySources({ sources }: { sources: StorySource[] }) {
  return (
    <section className="mt-12 p-6 sm:p-8 bg-slate-50 border border-slate-200 border-l-4 border-l-slate-950 rounded-r-3xl space-y-4">
      <div>
        <h2 className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-950">
          <BookMarked className="w-4 h-4" />
          <span>Источники</span>
        </h2>
        <p className="text-xs text-slate-500 mt-1">
          Материал проверен по архивным фондам ГАОО и верифицированным библиографическим источникам:
        </p>
      </div>
      <ul className="space-y-2.5 divide-y divide-slate-200/60 pt-2">
        {sources.map((src, i) => (
          <li key={i} className="pt-2.5 first:pt-0 flex items-start gap-3 text-xs sm:text-sm">
            <span className="px-2 py-0.5 bg-white border border-slate-200 rounded text-[10px] font-bold text-slate-700 uppercase shrink-0 mt-0.5">
              {src.typeLabel}
            </span>
            <span className="text-slate-800 font-mono leading-relaxed">{src.citation}</span>
          </li>
        ))}
      </ul>
    </section>
  );
}
