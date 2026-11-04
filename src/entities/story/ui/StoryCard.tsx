import Link from 'next/link';
import { ArrowRight } from 'lucide-react';
import { cn } from '@/shared/lib';
import type { Story } from '../model/types';

interface StoryCardProps {
  story: Story;
  compact?: boolean; // на главной: ниже обложка, короче анонс
}

export function StoryCard({ story, compact }: StoryCardProps) {
  return (
    <Link
      href={`/stories/${story.slug}`}
      className="group bg-white border border-slate-200 rounded-2xl overflow-hidden shadow-xs hover:shadow-md hover:border-slate-300 transition-all flex flex-col justify-between"
    >
      <div>
        <div className={cn('bg-slate-100 overflow-hidden relative', compact ? 'h-48' : 'h-52')}>
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={story.coverImage}
            alt={story.title}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
          />
          <span className="absolute top-3 left-3 px-2.5 py-1 bg-slate-950/85 backdrop-blur-md text-white text-[10px] font-bold uppercase tracking-wider rounded">
            {story.categoryLabel}
          </span>
        </div>
        <div className="p-5 space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400 font-medium">
            <span>{story.publishedAt}</span>
            <span>•</span>
            <span>{story.readTime}</span>
          </div>
          <h3 className="text-base font-bold text-slate-950 group-hover:text-blue-600 transition-colors leading-snug">
            {story.title}
          </h3>
          <p className={cn('text-xs text-slate-600 leading-relaxed', compact ? 'line-clamp-2' : 'line-clamp-3')}>
            {story.lead}
          </p>
        </div>
      </div>
      <div className="px-5 pb-5 pt-0 flex items-center justify-between text-xs font-semibold text-slate-900 group-hover:text-blue-600">
        <span className="text-[11px] text-slate-400 font-normal">{story.author.name}</span>
        <span className="flex items-center gap-1">
          <span>{compact ? 'Читать' : 'Читать статью'}</span>
          <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition-transform" />
        </span>
      </div>
    </Link>
  );
}
