import { Clock, Share2, User } from 'lucide-react';
import type { Story } from '@/entities/story';

// Шапка статьи: рубрика, дата, время чтения, заголовок, лид, автор (ТЗ 5.2).
export function StoryHeader({ story }: { story: Story }) {
  return (
    <header className="space-y-4 mb-8">
      <div className="flex flex-wrap items-center gap-2 text-xs">
        <span className="px-3 py-1 bg-slate-100 text-slate-900 font-bold uppercase tracking-wider rounded-md">
          {story.categoryLabel}
        </span>
        <span className="text-slate-300">•</span>
        <span className="text-slate-500 font-medium">{story.publishedAt}</span>
        <span className="text-slate-300">•</span>
        <span className="text-slate-500 font-medium flex items-center gap-1">
          <Clock className="w-3.5 h-3.5" />
          <span>{story.readTime}</span>
        </span>
      </div>

      <h1 className="text-3xl sm:text-5xl font-black text-slate-950 tracking-tight leading-tight">{story.title}</h1>
      <p className="text-base sm:text-xl text-slate-600 font-normal leading-relaxed">{story.lead}</p>

      <div className="flex items-center justify-between pt-4 border-t border-slate-100">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-slate-950 text-white flex items-center justify-center">
            <User className="w-5 h-5" />
          </div>
          <div>
            <div className="text-sm font-bold text-slate-950">{story.author.name}</div>
            <div className="text-xs text-slate-500">{story.author.role}</div>
          </div>
        </div>
        <button
          type="button"
          aria-label="Поделиться статьей"
          className="p-2 text-slate-500 hover:text-slate-950 rounded-lg border border-slate-200 transition-colors"
        >
          <Share2 className="w-4 h-4" />
        </button>
      </div>
    </header>
  );
}
