import Link from 'next/link';
import { Eye, MapPin } from 'lucide-react';
import { hasPanorama } from '../lib/hasPanorama';
import { objectMapHref, objectSubtitle } from '../lib/object';
import type { HeritageObject } from '../model/types';
import { ObjectCover } from './ObjectCover';
import { ProtectionBadge } from './ProtectionBadge';

// Карточка объекта в каталоге: обложка, период и стиль, адрес, статус охраны.
export function ObjectCard({ object }: { object: HeritageObject }) {
  return (
    <Link
      href={objectMapHref(object)}
      className="bg-white rounded-xl border border-slate-200 overflow-hidden shadow-xs hover:shadow-md hover:border-slate-300 transition-all flex flex-col justify-between group"
    >
      <div>
        <ObjectCover
          object={object}
          className={object.coverImage ? 'h-48 w-full' : 'h-32 w-full border-b border-slate-100'}
          imageClassName="group-hover:scale-105 transition-transform duration-300"
          iconClassName="w-8 h-8"
        >
          {hasPanorama(object) && (
            <div className="absolute top-3 right-3 bg-slate-900/85 backdrop-blur-md text-white text-[10px] font-bold px-2 py-0.5 rounded shadow flex items-center gap-1">
              <Eye className="w-3 h-3 text-blue-400" />
              <span>360°</span>
            </div>
          )}
        </ObjectCover>

        <div className="p-5">
          <div className="text-[11px] font-semibold text-blue-600 uppercase mb-2">
            {objectSubtitle(object)}
          </div>
          <h3 className="text-base font-bold text-slate-900 group-hover:text-blue-600 transition-colors mb-1.5 leading-snug">
            {object.title}
          </h3>
          <p className="text-xs text-slate-500 mb-2.5 flex items-center gap-1">
            <MapPin className="w-3.5 h-3.5 shrink-0 text-slate-400" />
            <span className="truncate">{object.address}</span>
          </p>
          <p className="text-xs text-slate-600 line-clamp-2 leading-relaxed">{object.shortDescription}</p>
        </div>
      </div>

      <div className="px-5 pb-5 pt-0">
        <div className="pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
          <ProtectionBadge status={object.protectionStatus} variant="text" />
          <span className="text-blue-600 font-semibold group-hover:translate-x-0.5 transition-transform">
            Смотреть на карте →
          </span>
        </div>
      </div>
    </Link>
  );
}
