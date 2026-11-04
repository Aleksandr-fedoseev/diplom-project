import Link from 'next/link';
import { MapPin } from 'lucide-react';
import { ObjectCover, objectMapHref, type HeritageObject } from '@/entities/heritage-object';

// Объекты карты, о которых говорится в статье (ТЗ 5.2).
export function RelatedObjects({ objects }: { objects: HeritageObject[] }) {
  return (
    <section className="mt-12 p-6 sm:p-8 bg-slate-50 border border-slate-200 rounded-3xl space-y-4">
      <h2 className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-950">
        <MapPin className="w-4 h-4 text-blue-600" />
        <span>Связанные объекты карты</span>
      </h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {objects.map((obj) => (
          <Link
            key={obj.id}
            href={objectMapHref(obj)}
            className="flex items-center gap-3 p-3.5 bg-white border border-slate-200 rounded-2xl hover:border-slate-400 transition-all group shadow-xs"
          >
            <ObjectCover object={obj} className="w-16 h-16 rounded-xl shrink-0" />
            <div className="min-w-0 flex-1">
              <span className="text-[10px] font-bold text-blue-600 uppercase">{obj.yearBuilt} г.</span>
              <h3 className="text-sm font-bold text-slate-950 group-hover:text-blue-600 transition-colors truncate">
                {obj.title}
              </h3>
              <p className="text-xs text-slate-500 truncate">{obj.address}</p>
              <span className="text-[11px] font-bold text-slate-900 group-hover:text-blue-600 block mt-1">
                Открыть на карте →
              </span>
            </div>
          </Link>
        ))}
      </div>
    </section>
  );
}
