'use client';

import { useState } from 'react';
import { BookOpen, Eye, MapPin, Share2, X } from 'lucide-react';
import {
  hasPanorama,
  ObjectCover,
  ObjectTags,
  ProtectionBadge,
  type HeritageObject,
} from '@/entities/heritage-object';
import { PanoramaViewer } from '@/entities/panorama';
import { storiesAboutObject, StoryLink } from '@/entities/story';
import { Tabs } from '@/shared/ui';

type DrawerTab = 'info' | 'panorama';

interface ObjectDrawerProps {
  object: HeritageObject;
  onClose: () => void;
}

// Панель объекта поверх карты: описание, панорама, статьи об объекте.
export default function ObjectDrawer({ object, onClose }: ObjectDrawerProps) {
  const [tab, setTab] = useState<DrawerTab>('info');
  const stories = storiesAboutObject(object.slug);

  return (
    <aside
      aria-label={object.title}
      className="absolute top-0 right-0 bottom-0 w-full sm:w-[460px] lg:w-[500px] bg-white border-l border-slate-200 shadow-2xl z-[1000] flex flex-col overflow-hidden animate-in slide-in-from-right duration-200"
    >
      <div className="p-4 sm:p-5 border-b border-slate-200 flex items-start justify-between gap-3 bg-slate-50/80">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <ProtectionBadge
              status={object.protectionStatus}
              className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5"
            />
            <span className="text-[11px] font-semibold text-slate-500">{object.yearBuilt} г.</span>
          </div>
          <h2 className="text-lg sm:text-xl font-bold text-slate-900 leading-snug">{object.title}</h2>
          <p className="text-xs text-slate-500 flex items-center gap-1 mt-1">
            <MapPin className="w-3.5 h-3.5 text-slate-400 shrink-0" />
            <span className="truncate">{object.address}</span>
          </p>
        </div>
        <button
          type="button"
          onClick={onClose}
          className="p-1.5 rounded-lg text-slate-400 hover:text-slate-700 hover:bg-slate-200 transition-colors shrink-0"
          aria-label="Закрыть"
        >
          <X className="w-5 h-5" />
        </button>
      </div>

      <Tabs
        size="sm"
        className="bg-white px-4 shrink-0"
        value={tab}
        onChange={setTab}
        items={[
          { id: 'info', label: 'Описание и история' },
          ...(hasPanorama(object)
            ? [{ id: 'panorama' as const, label: 'Панорама 360°', icon: <Eye className="w-3.5 h-3.5" /> }]
            : []),
        ]}
      />

      <div className="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
        {tab === 'info' && (
          <div className="space-y-5">
            {object.coverImage && (
              <ObjectCover object={object} className="h-52 sm:h-60 rounded-xl border border-slate-200 shadow-xs" />
            )}

            <dl className="grid grid-cols-2 gap-2.5 p-3.5 bg-slate-50 border border-slate-200 rounded-xl text-xs">
              <div>
                <dt className="text-slate-400 mb-0.5">Стиль:</dt>
                <dd className="font-semibold text-slate-900">{object.architecturalStyle}</dd>
              </div>
              {object.architect && (
                <div>
                  <dt className="text-slate-400 mb-0.5">Архитектор:</dt>
                  <dd className="font-semibold text-slate-900">{object.architect}</dd>
                </div>
              )}
            </dl>

            <section>
              <h3 className="text-sm font-bold text-slate-900 mb-1.5">Историческая справка</h3>
              <p className="text-sm text-slate-700 leading-relaxed">{object.fullHistory}</p>
            </section>

            <section>
              <h3 className="text-sm font-bold text-slate-900 mb-1.5">Архитектурные особенности</h3>
              <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">{object.shortDescription}</p>
            </section>

            {stories.length > 0 && (
              <section className="pt-4 border-t border-slate-200 space-y-3">
                <h3 className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-900">
                  <BookOpen className="w-4 h-4 text-blue-600" />
                  <span>Статьи и исследования об объекте:</span>
                </h3>
                <div className="space-y-2">
                  {stories.map((story) => (
                    <StoryLink key={story.id} story={story} />
                  ))}
                </div>
              </section>
            )}

            <ObjectTags tags={object.tags} className="pt-3 border-t border-slate-100 gap-1.5" tagClassName="px-2 py-0.5 text-[11px]" />
          </div>
        )}

        {tab === 'panorama' && <PanoramaViewer key={object.id} panoramas={object.panoramas ?? []} heightClassName="h-72 sm:h-80" />}
      </div>

      <div className="p-4 border-t border-slate-200 bg-slate-50 flex items-center justify-between gap-3 shrink-0">
        <span className="text-xs font-mono text-slate-500">
          GPS: {object.coordinates.lat.toFixed(4)}, {object.coordinates.lng.toFixed(4)}
        </span>
        <button
          type="button"
          onClick={() => {
            navigator.clipboard?.writeText(window.location.href);
            alert('Координаты и ссылка скопированы в буфер обмена!');
          }}
          className="px-3 py-1.5 bg-white hover:bg-slate-100 border border-slate-200 rounded-lg text-xs font-semibold text-slate-700 flex items-center gap-1.5 transition-colors"
        >
          <Share2 className="w-3.5 h-3.5" />
          <span>Поделиться</span>
        </button>
      </div>
    </aside>
  );
}
