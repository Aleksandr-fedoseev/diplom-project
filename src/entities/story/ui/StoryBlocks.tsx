/* eslint-disable @next/next/no-img-element */
import { Quote, Video, Volume2 } from 'lucide-react';
import { isTrustedVideoUrl } from '../lib/video';
import type { StoryContentBlock } from '../model/types';

// Текст статьи: абзацы, цитаты, галереи, аудио и видео (ТЗ 5.2).
export function StoryBlocks({ blocks }: { blocks: StoryContentBlock[] }) {
  return (
    <div className="space-y-8 text-base sm:text-lg text-slate-800 leading-relaxed">
      {blocks.map((block, idx) => (
        <StoryBlock key={idx} block={block} />
      ))}
    </div>
  );
}

function StoryBlock({ block }: { block: StoryContentBlock }) {
  switch (block.type) {
    case 'paragraph':
      return <p className="leading-relaxed">{block.text}</p>;

    case 'quote':
      return (
        <figure className="my-8 p-6 sm:p-8 bg-slate-50 border-l-4 border-slate-950 rounded-r-2xl space-y-2">
          <Quote className="w-8 h-8 text-slate-400 mb-2" />
          <blockquote className="text-lg sm:text-xl font-medium text-slate-900 italic leading-snug">{block.text}</blockquote>
          {block.author && (
            <figcaption className="text-xs sm:text-sm font-semibold text-slate-500 pt-2">— {block.author}</figcaption>
          )}
        </figure>
      );

    case 'image':
    case 'gallery':
      if (!block.images?.length) return null;
      return (
        <div className="my-8 grid grid-cols-1 sm:grid-cols-2 gap-4">
          {block.images.map((img, i) => (
            <figure key={i} className="rounded-2xl overflow-hidden border border-slate-200 bg-slate-50">
              <img src={img.url} alt={img.caption} className="w-full h-52 object-cover" />
              <figcaption className="p-3 text-xs text-slate-600">
                <strong className="block text-slate-900">{img.caption}</strong>
                {img.source && <span className="text-slate-400 text-[11px]">{img.source}</span>}
              </figcaption>
            </figure>
          ))}
        </div>
      );

    case 'audio':
      return (
        <div className="my-8 p-5 bg-slate-950 text-white rounded-2xl flex items-center justify-between gap-4">
          <div className="flex items-center gap-3 min-w-0">
            <div className="w-10 h-10 rounded-xl bg-white/10 flex items-center justify-center shrink-0">
              <Volume2 className="w-5 h-5 text-amber-400" />
            </div>
            <div className="min-w-0">
              <div className="text-[10px] text-slate-400 font-bold uppercase tracking-wider">Аудиоматериалы</div>
              <div className="text-sm font-bold truncate text-white">{block.audioTitle}</div>
            </div>
          </div>
          <span className="text-xs font-mono font-bold bg-white/10 px-2.5 py-1 rounded text-slate-200 shrink-0">
            {block.audioDuration}
          </span>
        </div>
      );

    case 'video':
      if (!isTrustedVideoUrl(block.videoUrl)) return null;
      return (
        <div className="my-8 border border-slate-200 rounded-2xl overflow-hidden bg-slate-900">
          <div className="p-4 bg-slate-950 text-white flex items-center gap-2">
            <Video className="w-4 h-4 text-rose-400" />
            <span className="text-xs font-bold uppercase tracking-wider">Видеоматериалы: {block.videoTitle}</span>
          </div>
          <div className="relative aspect-video bg-black">
            <iframe
              src={block.videoUrl}
              title={block.videoTitle || 'Видеоматериал'}
              className="w-full h-full border-0"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowFullScreen
            />
          </div>
        </div>
      );

    default:
      return null;
  }
}
