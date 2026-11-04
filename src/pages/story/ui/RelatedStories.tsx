import Link from 'next/link';
import { ArrowRight } from 'lucide-react';
import type { Story } from '@/entities/story';

// «Читайте также» (ТЗ 5.2).
export function RelatedStories({ stories }: { stories: Story[] }) {
  return (
    <section className="mt-12 pt-8 border-t border-slate-200 space-y-6">
      <h2 className="text-xl font-bold text-slate-950">Связанные статьи</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {stories.map((rel) => (
          <Link
            key={rel.id}
            href={`/stories/${rel.slug}`}
            className="p-5 bg-slate-50 border border-slate-200 rounded-2xl hover:border-slate-400 transition-all flex items-center justify-between group"
          >
            <div>
              <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">{rel.categoryLabel}</span>
              <h3 className="text-sm font-bold text-slate-950 group-hover:text-blue-600 transition-colors mt-0.5">{rel.title}</h3>
              <span className="text-xs text-slate-500 mt-1 block">{rel.readTime}</span>
            </div>
            <span className="inline-flex items-center gap-1 text-xs font-bold text-slate-900 group-hover:text-blue-600 shrink-0 ml-3">
              <span>Читать статью</span>
              <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition-transform" />
            </span>
          </Link>
        ))}
      </div>
    </section>
  );
}
