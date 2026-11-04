import Link from 'next/link';
import { Compass } from 'lucide-react';
import { ROUTES, routeStops, type RouteItem } from '@/entities/route';
import { plural } from '@/shared/lib';
import { PageHeader } from '@/shared/ui';

function RouteCard({ route }: { route: RouteItem }) {
  const stops = routeStops(route);
  return (
    <article id={route.slug} className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-xs hover:shadow-md transition-all space-y-4">
      <div className="flex flex-wrap items-center gap-3 text-xs font-semibold text-slate-500">
        <span className="px-2.5 py-1 bg-slate-100 rounded text-slate-800">Время: {route.duration}</span>
        <span>•</span>
        <span className="px-2.5 py-1 bg-slate-100 rounded text-slate-800">Дистанция: {route.distance}</span>
        <span>•</span>
        <span>
          {stops.length} {plural(stops.length, ['ключевой объект', 'ключевых объекта', 'ключевых объектов'])}
        </span>
      </div>

      <h2 className="text-2xl font-bold text-slate-900">{route.title}</h2>
      <p className="text-sm text-slate-600 leading-relaxed max-w-3xl">{route.description}</p>

      <div className="pt-3">
        <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-3">
          Последовательность памятников на маршруте:
        </h3>
        <ol className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {stops.map((obj, idx) => (
            <li key={obj.id}>
              <Link
                href={`/objects/${obj.slug}`}
                className="flex items-center gap-2.5 p-3 rounded-lg bg-slate-50 hover:bg-blue-50 hover:text-blue-700 transition-colors border border-slate-200 text-xs font-medium text-slate-800"
              >
                <span className="w-5 h-5 rounded-full bg-slate-900 text-white flex items-center justify-center font-bold text-[10px] shrink-0">
                  {idx + 1}
                </span>
                <span className="truncate font-semibold">{obj.title}</span>
              </Link>
            </li>
          ))}
        </ol>
      </div>

      <div className="pt-4 border-t border-slate-100">
        <Link
          href="/map"
          className="inline-flex items-center gap-2 px-5 py-2.5 bg-slate-900 hover:bg-blue-600 text-white rounded-lg text-xs font-semibold transition-colors"
        >
          <Compass className="w-4 h-4" />
          <span>Показать объекты на карте</span>
        </Link>
      </div>
    </article>
  );
}

export default function RoutesPage() {
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-14 w-full">
      <PageHeader
        className="mb-10"
        kicker="Туристическая навигация"
        title="Историко-архитектурные маршруты"
        lead="Пешеходные экскурсии по историческому центру Астрахани с точными координатами памятников на карте OpenStreetMap."
      />
      <div className="space-y-6">
        {ROUTES.map((route) => (
          <RouteCard key={route.id} route={route} />
        ))}
      </div>
    </div>
  );
}
