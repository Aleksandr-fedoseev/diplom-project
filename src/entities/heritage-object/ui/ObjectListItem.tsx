import { MapPin } from 'lucide-react';
import { cn } from '@/shared/lib';
import { hasPanorama } from '../lib/hasPanorama';
import type { HeritageObject } from '../model/types';
import { ObjectCover } from './ObjectCover';
import { ProtectionBadge } from './ProtectionBadge';

interface ObjectListItemProps {
  object: HeritageObject;
  selected: boolean;
  onSelect: (object: HeritageObject) => void;
}

// Строка списка объектов рядом с картой.
export function ObjectListItem({ object, selected, onSelect }: ObjectListItemProps) {
  return (
    <button
      type="button"
      onClick={() => onSelect(object)}
      aria-current={selected || undefined}
      className={cn(
        'w-full text-left p-3.5 cursor-pointer transition-colors flex gap-3',
        selected ? 'bg-blue-50/80 border-l-4 border-blue-600' : 'hover:bg-slate-50',
      )}
    >
      <ObjectCover object={object} className="w-16 h-16 rounded-lg shrink-0 border border-slate-200" />

      <div className="flex-1 min-w-0 flex flex-col justify-between">
        <div>
          <div className="flex items-center justify-between gap-1">
            <span className="text-[10px] font-semibold text-blue-600 uppercase truncate">
              {object.yearBuilt} г. • {object.architecturalStyle}
            </span>
            {hasPanorama(object) && (
              <span className="bg-blue-600 text-white text-[8px] font-bold px-1 rounded shrink-0">360°</span>
            )}
          </div>
          <h3 className="text-xs sm:text-sm font-bold text-slate-900 leading-snug truncate mt-0.5">{object.title}</h3>
          <p className="text-[11px] text-slate-500 truncate flex items-center gap-1 mt-0.5">
            <MapPin className="w-3 h-3 text-slate-400 shrink-0" />
            <span>{object.address}</span>
          </p>
        </div>

        <div className="flex items-center justify-between pt-1 text-[10px]">
          <ProtectionBadge status={object.protectionStatus} length="short" className="px-1.5 py-0.5" />
          <span className="font-semibold text-blue-600">Подробнее →</span>
        </div>
      </div>
    </button>
  );
}
