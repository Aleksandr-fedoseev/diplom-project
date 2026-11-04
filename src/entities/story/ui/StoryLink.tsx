import Link from 'next/link';
import { ArrowRight } from 'lucide-react';
import type { Story } from '../model/types';

// Компактная ссылка на статью: рубрика, заголовок, «Читать».
export function StoryLink({ story }: { story: Story }) {
  return (
    <Link
      href={`/stories/${story.slug}`}
      className="p-3 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-xl flex items-center justify-between group transition-colors"
    >
      <div className="min-w-0 flex-1 pr-2">
        <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider block">{story.categoryLabel}</span>
        <h4 className="text-xs sm:text-sm font-bold text-slate-950 group-hover:text-blue-600 transition-colors truncate">
          {story.title}
        </h4>
      </div>
      <span className="text-xs font-bold text-slate-900 group-hover:text-blue-600 flex items-center gap-1 shrink-0">
        <span>Читать</span>
        <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-0.5 transition-transform" />
      </span>
    </Link>
  );
}
