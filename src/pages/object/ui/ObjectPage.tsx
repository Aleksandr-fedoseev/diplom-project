'use client';

import { useState } from 'react';
import Link from 'next/link';
import { notFound, useParams } from 'next/navigation';
import { ApiError } from '@/shared/api';
import { ArrowLeft, Eye, FileText, MapPin, Share2 } from 'lucide-react';
import { hasPanorama, ObjectTags, ProtectionBadge, useObject } from '@/entities/heritage-object';
import { PanoramaViewer } from '@/entities/panorama';
import { ErrorMessage, Loading, Tabs } from '@/shared/ui';
import { ObjectFacts } from './ObjectFacts';

type ObjectTab = 'passport' | 'panorama';

export default function ObjectPage() {
  const slug = useParams<{ slug: string }>()?.slug ?? '';
  const card = useObject(slug);
  const [tab, setTab] = useState<ObjectTab>('passport');

  if (card.error instanceof ApiError && card.error.status === 404) notFound();
  if (card.error)
    return (
      <div className="max-w-7xl mx-auto px-4 py-12">
        <ErrorMessage error={card.error} onRetry={() => card.refetch()} />
      </div>
    );
  if (!card.data) return <Loading label="Загрузка паспорта объекта..." />;
  const object = card.data;

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12 w-full">
      <div className="flex items-center justify-between gap-4 mb-6">
        <Link
          href="/objects"
          className="inline-flex items-center gap-1.5 text-sm font-semibold text-slate-600 hover:text-slate-900 transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Вернуться в каталог</span>
        </Link>
        <button
          type="button"
          onClick={() => {
            navigator.clipboard?.writeText(window.location.href);
            alert('Ссылка на паспорт скопирована!');
          }}
          className="p-2 bg-white rounded-lg border border-slate-200 text-slate-600 hover:text-slate-900 hover:bg-slate-50 transition-colors text-xs flex items-center gap-1.5"
          title="Поделиться"
        >
          <Share2 className="w-4 h-4" />
          <span className="hidden sm:inline">Поделиться</span>
        </button>
      </div>

      <div className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-xs mb-8">
        <div className="flex flex-wrap items-center gap-2 mb-3">
          <ProtectionBadge
            status={object.protectionStatus}
            length="long"
            variant="outlined"
            className="text-xs font-bold uppercase tracking-wider px-2.5 py-1"
          />
          {object.registryNumber && (
            <span className="text-xs font-mono font-medium text-slate-500 bg-slate-100 px-2.5 py-1 rounded">
              № {object.registryNumber}
            </span>
          )}
        </div>
        <h1 className="text-3xl sm:text-4xl font-bold text-slate-900 tracking-tight mb-2">{object.title}</h1>
        <p className="text-sm sm:text-base text-slate-600 flex items-center gap-1.5 mb-6">
          <MapPin className="w-4 h-4 text-slate-400 shrink-0" />
          <span>{object.address}</span>
        </p>

        <Tabs
          value={tab}
          onChange={setTab}
          items={[
            { id: 'passport', label: 'Паспорт и историческая справка', icon: <FileText className="w-4 h-4" /> },
            ...(hasPanorama(object)
              ? [
                  {
                    id: 'panorama' as const,
                    icon: <Eye className="w-4 h-4" />,
                    label: (
                      <>
                        <span>Сферическая панорама 360°</span>
                        <span className="text-[10px] bg-blue-100 text-blue-800 px-1.5 py-0.5 rounded font-bold">VR</span>
                      </>
                    ),
                  },
                ]
              : []),
          ]}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        <div className="lg:col-span-8 space-y-6">
          {tab === 'passport' && (
            <div className="bg-white rounded-2xl p-6 sm:p-8 border border-slate-200 shadow-xs space-y-6">
              <section>
                <h2 className="text-xl font-bold text-slate-900 mb-3">Архитектурно-историческое описание</h2>
                <p className="text-base text-slate-700 leading-relaxed">
                  {object.fullHistory || 'Историческая справка пока не добавлена.'}
                </p>
              </section>
              <section className="pt-4 border-t border-slate-100">
                <h3 className="text-base font-bold text-slate-900 mb-2">Краткая характеристика</h3>
                <p className="text-sm text-slate-600 leading-relaxed">{object.shortDescription}</p>
              </section>
              <ObjectTags tags={object.tags} className="pt-4 border-t border-slate-100" />
            </div>
          )}
          {tab === 'panorama' && <PanoramaViewer key={object.id} panoramas={object.panoramas ?? []} />}
        </div>

        <div className="lg:col-span-4 space-y-6">
          <ObjectFacts object={object} />
        </div>
      </div>
    </div>
  );
}
