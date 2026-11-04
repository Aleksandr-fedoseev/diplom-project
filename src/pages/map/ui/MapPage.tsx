'use client';

import { Suspense, useCallback, useEffect, useState } from 'react';
import dynamic from 'next/dynamic';
import { useSearchParams } from 'next/navigation';
import { findObjectBySlug, HERITAGE_OBJECTS, ObjectListItem, type HeritageObject } from '@/entities/heritage-object';
import { FilterChips, useObjectFilter } from '@/features/filter-objects';
import { cn } from '@/shared/lib';
import { SearchInput } from '@/shared/ui';
import ObjectDrawer from './ObjectDrawer';

const Spinner = () => <div className="w-8 h-8 border-3 border-blue-600 border-t-transparent rounded-full animate-spin" />;

// Leaflet работает только в браузере, поэтому карта загружается без серверного рендеринга.
const HeritageMap = dynamic(() => import('./HeritageMap'), {
  ssr: false,
  loading: () => (
    <div className="w-full h-full bg-slate-100 flex flex-col items-center justify-center text-slate-500 gap-3">
      <Spinner />
      <span className="text-xs font-semibold">Загрузка OpenStreetMap Астраханской области...</span>
    </div>
  ),
});

function MapView() {
  const objectSlug = useSearchParams()?.get('object') ?? null;
  const { filter, update, filtered } = useObjectFilter(HERITAGE_OBJECTS);
  const [active, setActive] = useState<HeritageObject | null>(null);
  // На телефоне список и карта не помещаются рядом — показываем что-то одно.
  const [mobileView, setMobileView] = useState<'list' | 'map'>('list');
  const select = useCallback((o: HeritageObject) => setActive(o), []);

  // /map?object=dom-kononova — сразу открыть объект.
  useEffect(() => {
    if (objectSlug) setActive(findObjectBySlug(objectSlug) ?? null);
  }, [objectSlug]);

  return (
    <div className="h-[calc(100vh-4rem)] sm:h-[calc(100vh-5rem)] flex flex-col bg-slate-100 overflow-hidden relative">
      <div className="bg-white border-b border-slate-200 px-4 py-3 z-20 flex flex-wrap items-center justify-between gap-3 shadow-xs">
        <SearchInput
          className="flex-1 min-w-[220px] max-w-md"
          placeholder="Поиск по памятникам Астрахани..."
          value={filter.query}
          onChange={(query) => update({ query })}
        />
        <FilterChips filter={filter} onChange={update} />
        <div className="text-xs font-medium text-slate-500 hidden xl:flex items-center gap-1.5 bg-slate-100 px-2.5 py-1.5 rounded-lg border border-slate-200">
          <span className="w-2 h-2 rounded-full bg-emerald-500" />
          <span>Карта OpenStreetMap (Астраханская область)</span>
        </div>
      </div>

      <div role="tablist" aria-label="Вид" className="md:hidden grid grid-cols-2 gap-1 p-1 m-2 bg-slate-200/80 rounded-lg text-xs font-semibold">
        {(['list', 'map'] as const).map((v) => (
          <button
            key={v}
            type="button"
            role="tab"
            aria-selected={mobileView === v}
            onClick={() => setMobileView(v)}
            className={cn('py-1.5 rounded-md', mobileView === v ? 'bg-white text-slate-900 shadow-xs' : 'text-slate-600')}
          >
            {v === 'list' ? `Список (${filtered.length})` : 'Карта'}
          </button>
        ))}
      </div>

      <div className="flex-1 flex overflow-hidden relative">
        <div
          className={cn(
            'w-full md:w-[340px] lg:w-[380px] bg-white border-r border-slate-200 flex-col shrink-0 overflow-hidden z-10',
            mobileView === 'list' ? 'flex' : 'hidden md:flex',
          )}
        >
          <div className="p-3 border-b border-slate-100 bg-slate-50 text-xs font-medium text-slate-600 flex justify-between items-center">
            <span>Объектов: {filtered.length}</span>
            <span className="text-slate-400">Кликните для просмотра</span>
          </div>
          <div className="flex-1 overflow-y-auto divide-y divide-slate-100">
            {filtered.length === 0 ? (
              <div className="p-8 text-center text-slate-500 text-sm">По заданному фильтру объектов не найдено.</div>
            ) : (
              filtered.map((o) => (
                <ObjectListItem key={o.id} object={o} selected={active?.id === o.id} onSelect={select} />
              ))
            )}
          </div>
        </div>

        <div className={cn('flex-1 relative overflow-hidden bg-slate-200', mobileView === 'list' && 'hidden md:block')}>
          <HeritageMap objects={filtered} activeObject={active} onSelectObject={select} />
        </div>

        {active && <ObjectDrawer key={active.id} object={active} onClose={() => setActive(null)} />}
      </div>
    </div>
  );
}

export default function MapPage() {
  return (
    <Suspense
      fallback={
        <div className="h-[calc(100vh-4rem)] sm:h-[calc(100vh-5rem)] flex items-center justify-center bg-slate-100">
          <Spinner />
        </div>
      }
    >
      <MapView />
    </Suspense>
  );
}
