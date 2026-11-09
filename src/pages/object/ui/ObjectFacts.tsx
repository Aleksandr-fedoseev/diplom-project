import Link from 'next/link';
import { MapPin } from 'lucide-react';
import { formatCoordinates, objectMapHref, type HeritageObject } from '@/entities/heritage-object';

// Паспортные реквизиты объекта в боковой колонке.
export function ObjectFacts({ object }: { object: HeritageObject }) {
  const facts: [string, string | undefined][] = [
    ['Период постройки', object.yearBuilt && `${object.yearBuilt} г.`],
    ['Архитектурный стиль', object.architecturalStyle],
    ['Архитектор / Зодчий', object.architect],
    ['Адрес объекта', object.address],
  ];

  return (
    <div className="bg-white rounded-2xl p-6 border border-slate-200 shadow-xs space-y-5">
      <h3 className="text-base font-bold text-slate-900 border-b border-slate-100 pb-3">Паспортные реквизиты</h3>

      <dl className="space-y-3.5 text-sm">
        {facts
          .filter(([, value]) => value)
          .map(([label, value]) => (
            <div key={label}>
              <dt className="text-xs text-slate-400">{label}</dt>
              <dd className="font-semibold text-slate-900">{value}</dd>
            </div>
          ))}
        <div>
          <dt className="text-xs text-slate-400">Географические координаты (WGS 84)</dt>
          <dd className="font-mono text-xs text-slate-700">{formatCoordinates(object.coordinates)}</dd>
        </div>
      </dl>

      <div className="pt-3 border-t border-slate-100">
        <Link
          href={objectMapHref(object)}
          className="w-full inline-flex items-center justify-center gap-2 py-2.5 px-4 bg-slate-900 hover:bg-blue-600 text-white rounded-lg text-xs font-semibold transition-colors"
        >
          <MapPin className="w-4 h-4 text-blue-400" />
          <span>Показать на интерактивной карте</span>
        </Link>
        <Link
          href={`/contribute?object=${encodeURIComponent(object.slug)}`}
          className="mt-2 w-full inline-flex items-center justify-center py-2.5 px-4 bg-slate-100 hover:bg-slate-200 text-slate-800 rounded-lg text-xs font-semibold transition-colors"
        >
          Дополнить: фото, архив, уточнения
        </Link>
      </div>
    </div>
  );
}
