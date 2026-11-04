/* eslint-disable @next/next/no-img-element */
import Link from 'next/link';
import { ArrowRight, Clock, Map as MapIcon, UserCheck } from 'lucide-react';
import type { Tour } from '@/entities/tour';
import { formatRub, plural } from '@/shared/lib';

interface TourCardProps {
  tour: Tour;
  onBook: (tourId: string) => void;
}

export function TourCard({ tour, onBook }: TourCardProps) {
  const stops = tour.stops.length;
  return (
    <article
      id={tour.slug}
      className="bg-white border border-slate-200 rounded-3xl overflow-hidden shadow-xs hover:shadow-md transition-all grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch"
    >
      <div className="lg:col-span-5 h-64 lg:h-auto bg-slate-100 relative">
        <img src={tour.coverImage} alt={tour.title} className="w-full h-full object-cover" />
        <span className="absolute top-4 left-4 px-3 py-1 bg-slate-950/85 backdrop-blur-md text-white text-xs font-bold uppercase tracking-wider rounded-lg">
          {tour.isOnlineRoute ? 'Онлайн-маршрут' : 'Пешеходная группа'}
        </span>
      </div>

      <div className="lg:col-span-7 p-6 sm:p-8 flex flex-col justify-between space-y-6">
        <div className="space-y-4">
          <div className="flex flex-wrap items-center gap-3 text-xs font-semibold text-slate-500">
            <span className="flex items-center gap-1.5 bg-slate-100 px-2.5 py-1 rounded text-slate-800">
              <Clock className="w-3.5 h-3.5" />
              <span>{tour.duration}</span>
            </span>
            <span>•</span>
            <span>{tour.distance}</span>
            <span>•</span>
            <span>
              {stops} {plural(stops, ['остановка', 'остановки', 'остановок'])}
            </span>
            {tour.nextDate && (
              <>
                <span>•</span>
                <span className="text-slate-900 font-bold bg-amber-50 border border-amber-200 px-2 py-0.5 rounded">
                  Ближайшая: {tour.nextDate}
                </span>
              </>
            )}
          </div>

          <h2 className="text-2xl font-bold text-slate-950 leading-snug">{tour.title}</h2>
          <p className="text-sm text-slate-600 leading-relaxed">{tour.description}</p>

          <div className="pt-2">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-2">Ключевые точки маршрута:</h3>
            <ol className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
              {tour.stops.map((stop, i) => (
                <li key={i} className="flex items-start gap-2 p-2 bg-slate-50 rounded-lg border border-slate-100">
                  <span className="w-4 h-4 rounded-full bg-slate-900 text-white flex items-center justify-center font-bold text-[9px] shrink-0 mt-0.5">
                    {i + 1}
                  </span>
                  <div>
                    <strong className="block text-slate-900 leading-tight">{stop.title}</strong>
                    <span className="text-slate-500 text-[10px]">{stop.address}</span>
                  </div>
                </li>
              ))}
            </ol>
          </div>

          {tour.guideName && (
            <div className="pt-2 text-xs text-slate-500 flex items-center gap-2">
              <UserCheck className="w-4 h-4 text-slate-700" />
              <span>
                Ведущий: <strong>{tour.guideName}</strong> ({tour.guideRole})
              </span>
            </div>
          )}
        </div>

        <div className="pt-4 border-t border-slate-100 flex flex-wrap items-center justify-between gap-4">
          <div>
            <span className="text-[11px] text-slate-400 uppercase tracking-wider block font-bold">Стоимость:</span>
            <span className="text-2xl font-black text-slate-950">{tour.price === 0 ? 'Бесплатно' : formatRub(tour.price)}</span>
          </div>
          {tour.isOnlineRoute ? (
            <Link
              href="/map"
              className="inline-flex items-center gap-2 px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-xl text-xs font-bold uppercase tracking-wider transition-colors shadow-sm"
            >
              <MapIcon className="w-4 h-4" />
              <span>Открыть маршрут на карте</span>
            </Link>
          ) : (
            <a
              href="#booking"
              onClick={() => onBook(tour.id)}
              className="inline-flex items-center gap-2 px-6 py-3 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-bold uppercase tracking-wider transition-colors shadow-xs"
            >
              <span>Записаться в группу</span>
              <ArrowRight className="w-4 h-4" />
            </a>
          )}
        </div>
      </div>
    </article>
  );
}
