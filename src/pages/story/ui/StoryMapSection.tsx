import Link from 'next/link';
import { ArrowRight, Compass } from 'lucide-react';
import type { Story } from '@/entities/story';

// Место, о котором статья, со ссылкой на карту (ТЗ 5.2).
export function StoryMapSection({ place, objectSlug }: { place: NonNullable<Story['mapCoordinates']>; objectSlug?: string }) {
  return (
    <section className="mt-12 p-6 bg-slate-50 border border-slate-200 rounded-3xl space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-950">
          <Compass className="w-4 h-4 text-blue-600" />
          <span>Карта объекта</span>
        </h2>
        <Link
          href={objectSlug ? `/map?object=${encodeURIComponent(objectSlug)}` : '/map'}
          className="text-xs font-bold text-slate-900 hover:text-blue-600 flex items-center gap-1 transition-colors"
        >
          <span>Открыть на карте</span>
          <ArrowRight className="w-3.5 h-3.5" />
        </Link>
      </div>
      <div className="bg-white p-4 rounded-xl border border-slate-200 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h3 className="text-sm font-bold text-slate-950">{place.title}</h3>
          <p className="text-xs text-slate-500 mt-0.5">{place.address}</p>
        </div>
        <div className="text-xs font-mono bg-slate-100 px-3 py-1.5 rounded text-slate-700 self-start sm:self-auto">
          {place.lat.toFixed(4)}° N, {place.lng.toFixed(4)}° E
        </div>
      </div>
    </section>
  );
}
