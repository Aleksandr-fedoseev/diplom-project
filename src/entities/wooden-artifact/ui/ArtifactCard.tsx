/* eslint-disable @next/next/no-img-element */
import type { WoodenArtifact } from '../model/types';

// Экспонат музея деревянной архитектуры.
export function ArtifactCard({ item }: { item: WoodenArtifact }) {
  return (
    <article className="bg-white border border-slate-200 rounded-2xl overflow-hidden shadow-xs hover:shadow-md transition-all flex flex-col justify-between group">
      <div>
        <div className="h-56 bg-slate-100 relative overflow-hidden">
          <img src={item.coverImage} alt={item.title} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
          <span className="absolute top-3 left-3 px-2.5 py-1 bg-slate-950/85 backdrop-blur-md text-white text-[10px] font-bold uppercase tracking-wider rounded">
            {item.categoryLabel}
          </span>
          <span className="absolute bottom-3 right-3 px-2.5 py-1 bg-white/90 backdrop-blur-md text-slate-900 text-[10px] font-bold rounded shadow-xs">
            {item.statusLabel}
          </span>
        </div>
        <div className="p-5 space-y-2.5">
          <h3 className="text-base font-bold text-slate-950 leading-snug">{item.title}</h3>
          <div className="text-xs text-slate-500 space-y-1">
            <div className="font-medium text-slate-700">{item.originalAddress}</div>
            <div className="text-[11px] text-slate-400">Период создания: {item.yearBuilt}</div>
          </div>
          <p className="text-xs text-slate-600 leading-relaxed pt-2 border-t border-slate-100">{item.description}</p>
        </div>
      </div>
      <div className="px-5 pb-5 pt-0 text-[11px] text-slate-400 font-mono flex items-center justify-between">
        <span>Год спасения:</span>
        <span className="font-bold text-slate-700">{item.acquisitionYear}</span>
      </div>
    </article>
  );
}
