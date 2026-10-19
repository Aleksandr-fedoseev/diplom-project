import React from 'react';
import Link from 'next/link';
import { Compass, Clock, Footprints, Landmark, ArrowRight, MapPin } from 'lucide-react';
import { ROUTES, HERITAGE_OBJECTS } from '@/data/heritage';

export default function RoutesPage() {
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-14 w-full">
      
      {/* Header */}
      <div className="mb-10">
        <div className="text-xs font-bold uppercase tracking-wider text-blue-600 mb-1">
          Туристическая навигация
        </div>
        <h1 className="text-3xl sm:text-4xl font-bold text-slate-900 tracking-tight">
          Историко-архитектурные маршруты
        </h1>
        <p className="text-base text-slate-600 mt-2 max-w-2xl">
          Пешеходные экскурсии по историческому центру Астрахани с точными координатами памятников на карте OpenStreetMap.
        </p>
      </div>

      {/* Routes List */}
      <div className="space-y-6">
        {ROUTES.map((route) => {
          const routeObjects = HERITAGE_OBJECTS.filter((o) => route.objectIds.includes(o.id));

          return (
            <div
              key={route.id}
              id={route.slug}
              className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-xs hover:shadow-md transition-all"
            >
              <div className="space-y-4">
                <div className="flex flex-wrap items-center gap-3 text-xs font-semibold text-slate-500">
                  <span className="px-2.5 py-1 bg-slate-100 rounded text-slate-800">
                    Время: {route.duration}
                  </span>
                  <span>•</span>
                  <span className="px-2.5 py-1 bg-slate-100 rounded text-slate-800">
                    Дистанция: {route.distance}
                  </span>
                  <span>•</span>
                  <span>{routeObjects.length} ключевых объекта</span>
                </div>

                <h2 className="text-2xl font-bold text-slate-900">
                  {route.title}
                </h2>

                <p className="text-sm text-slate-600 leading-relaxed max-w-3xl">
                  {route.description}
                </p>

                {/* Stops sequence */}
                <div className="pt-3">
                  <h4 className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-3">
                    Последовательность памятников на маршруте:
                  </h4>
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                    {routeObjects.map((obj, idx) => (
                      <Link
                        key={obj.id}
                        href={`/objects/${obj.slug}`}
                        className="flex items-center gap-2.5 p-3 rounded-lg bg-slate-50 hover:bg-blue-50 hover:text-blue-700 transition-colors border border-slate-200 text-xs font-medium text-slate-800"
                      >
                        <span className="w-5 h-5 rounded-full bg-slate-900 text-white flex items-center justify-center font-bold text-[10px] shrink-0">
                          {idx + 1}
                        </span>
                        <span className="truncate font-semibold">{obj.title}</span>
                      </Link>
                    ))}
                  </div>
                </div>

                <div className="pt-4 border-t border-slate-100 flex items-center gap-3">
                  <Link
                    href="/map"
                    className="inline-flex items-center gap-2 px-5 py-2.5 bg-slate-900 hover:bg-blue-600 text-white rounded-lg text-xs font-semibold transition-colors"
                  >
                    <Compass className="w-4 h-4" />
                    <span>Показать объекты на карте</span>
                  </Link>
                </div>

              </div>
            </div>
          );
        })}
      </div>

    </div>
  );
}
