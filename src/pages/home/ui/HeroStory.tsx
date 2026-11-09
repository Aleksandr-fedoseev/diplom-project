import Link from 'next/link';
import { ArrowRight, Compass } from 'lucide-react';
import type { Story } from '@/entities/story';

// Главный материал на первом экране.
export function HeroStory({ story }: { story: Story }) {
  const mapHref = story.relatedObjectSlugs?.[0] ? `/map?object=${encodeURIComponent(story.relatedObjectSlugs[0])}` : '/map';
  return (
    <section className="border-b border-slate-200 bg-slate-50/50 pt-6 pb-10 sm:pt-8 sm:pb-12">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="bg-ink-950 rounded-3xl text-snow overflow-hidden relative shadow-lg min-h-[400px] sm:min-h-[460px] p-6 sm:p-10 lg:p-12 flex flex-col justify-between">
          <div
            className="absolute inset-0 bg-cover bg-center opacity-40 mix-blend-luminosity"
            style={{ backgroundImage: `url(${story.coverImage})` }}
          />
          <div className="absolute inset-0 bg-gradient-to-t from-ink-950 via-ink-950/60 to-transparent" />

          <div className="relative z-10 space-y-4 max-w-3xl">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded bg-snow/20 backdrop-blur-md text-xs font-bold uppercase tracking-wider text-snow">
              <span>{story.categoryLabel}</span>
              <span>•</span>
              <span>{story.publishedAt}</span>
              <span>•</span>
              <span>{story.readTime}</span>
            </div>
            <h1 className="text-2xl sm:text-4xl lg:text-5xl font-black tracking-tight leading-tight text-snow">{story.title}</h1>
            <p className="text-sm sm:text-base lg:text-lg text-ink-200 leading-relaxed font-normal">{story.lead}</p>
          </div>

          <div className="relative z-10 pt-6 flex flex-wrap items-center gap-4">
            <Link
              href={`/stories/${story.slug}`}
              className="inline-flex items-center gap-2 px-6 py-3.5 bg-snow hover:bg-ink-100 text-ink-950 text-xs sm:text-sm font-bold uppercase tracking-wider rounded-xl transition-all shadow-md"
            >
              <span>Читать статью</span>
              <ArrowRight className="w-4 h-4" />
            </Link>
            <Link
              href={mapHref}
              className="inline-flex items-center gap-2 px-5 py-3.5 bg-ink-800/80 hover:bg-ink-800 text-snow text-xs sm:text-sm font-semibold rounded-xl backdrop-blur-md transition-all border border-ink-700"
            >
              <Compass className="w-4 h-4" />
              <span>Открыть на карте</span>
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}
