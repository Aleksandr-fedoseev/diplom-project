'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { 
  X, 
  Eye, 
  Volume2, 
  MapPin, 
  Calendar, 
  ShieldCheck,
  Share2,
  BookOpen,
  ArrowRight
} from 'lucide-react';
import { HeritageObject } from '@/types/heritage';
import { STORIES } from '@/data/stories';
import { hasPanorama } from '@/data/heritage';
import PanoramaViewer from '@/components/panorama/PanoramaViewer';

interface ObjectDetailDrawerProps {
  object: HeritageObject | null;
  onClose: () => void;
}

export default function ObjectDetailDrawer({ object, onClose }: ObjectDetailDrawerProps) {
  const [activeTab, setActiveTab] = useState<'info' | 'panorama' | 'audio'>('info');
  const [isPlayingAudio, setIsPlayingAudio] = useState(false);

  if (!object) return null;

  return (
    <div className="absolute top-0 right-0 bottom-0 w-full sm:w-[460px] lg:w-[500px] bg-white border-l border-slate-200 shadow-2xl z-[1000] flex flex-col overflow-hidden animate-in slide-in-from-right duration-200">
      
      {/* Drawer Header */}
      <div className="p-4 sm:p-5 border-b border-slate-200 flex items-start justify-between gap-3 bg-slate-50/80">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded ${
              object.protectionStatus === 'federal'
                ? 'bg-emerald-100 text-emerald-800'
                : object.protectionStatus === 'lost'
                ? 'bg-rose-100 text-rose-800'
                : 'bg-slate-200 text-slate-800'
            }`}>
              {object.protectionStatus === 'federal' && 'Федеральный ОКН'}
              {object.protectionStatus === 'regional' && 'Региональный ОКН'}
              {object.protectionStatus === 'lost' && 'Утраченный памятник'}
            </span>
            <span className="text-[11px] font-semibold text-slate-500">
              {object.yearBuilt} г.
            </span>
          </div>

          <h2 className="text-lg sm:text-xl font-bold text-slate-900 leading-snug">
            {object.title}
          </h2>

          <p className="text-xs text-slate-500 flex items-center gap-1 mt-1">
            <MapPin className="w-3.5 h-3.5 text-slate-400 shrink-0" />
            <span className="truncate">{object.address}</span>
          </p>
        </div>

        <button
          onClick={onClose}
          className="p-1.5 rounded-lg text-slate-400 hover:text-slate-700 hover:bg-slate-200 transition-colors shrink-0"
          aria-label="Закрыть"
        >
          <X className="w-5 h-5" />
        </button>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-slate-200 bg-white px-4 shrink-0 text-xs font-semibold">
        <button
          onClick={() => setActiveTab('info')}
          className={`py-3 px-3 border-b-2 transition-colors ${
            activeTab === 'info'
              ? 'border-blue-600 text-blue-600 font-bold'
              : 'border-transparent text-slate-600 hover:text-slate-900'
          }`}
        >
          Описание и история
        </button>

        {hasPanorama(object) && (
          <button
            onClick={() => setActiveTab('panorama')}
            className={`py-3 px-3 border-b-2 flex items-center gap-1.5 transition-colors ${
              activeTab === 'panorama'
                ? 'border-blue-600 text-blue-600 font-bold'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Eye className="w-3.5 h-3.5" />
            <span>Панорама 360°</span>
          </button>
        )}

        {object.hasAudioGuide && (
          <button
            onClick={() => setActiveTab('audio')}
            className={`py-3 px-3 border-b-2 flex items-center gap-1.5 transition-colors ${
              activeTab === 'audio'
                ? 'border-blue-600 text-blue-600 font-bold'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Volume2 className="w-3.5 h-3.5" />
            <span>Аудиогид</span>
          </button>
        )}
      </div>

      {/* Scrollable Content */}
      <div className="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
        
        {/* TAB 1: Information */}
        {activeTab === 'info' && (
          <div className="space-y-5">
            {object.coverImage && (
              <div className="h-52 sm:h-60 rounded-xl overflow-hidden bg-slate-100 border border-slate-200 shadow-xs relative">
                <img
                  src={object.coverImage}
                  alt={object.title}
                  className="w-full h-full object-cover"
                />
              </div>
            )}

            {/* Quick Specs Grid */}
            <div className="grid grid-cols-2 gap-2.5 p-3.5 bg-slate-50 border border-slate-200 rounded-xl text-xs">
              <div>
                <span className="text-slate-400 block mb-0.5">Стиль:</span>
                <span className="font-semibold text-slate-900">{object.architecturalStyle}</span>
              </div>
              {object.architect && (
                <div>
                  <span className="text-slate-400 block mb-0.5">Архитектор:</span>
                  <span className="font-semibold text-slate-900">{object.architect}</span>
                </div>
              )}
            </div>

            {/* Description */}
            <div>
              <h3 className="text-sm font-bold text-slate-900 mb-1.5">
                Историческая справка
              </h3>
              <p className="text-sm text-slate-700 leading-relaxed">
                {object.fullHistory}
              </p>
            </div>

            {/* Additional info */}
            <div>
              <h3 className="text-sm font-bold text-slate-900 mb-1.5">
                Архитектурные особенности
              </h3>
              <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
                {object.shortDescription}
              </p>
            </div>

            {/* 7.1 / 7.3 Связь карты и статей: Связанные статьи об объекте */}
            {STORIES.filter(s => s.relatedObjectSlugs?.includes(object.slug)).length > 0 && (
              <div className="pt-4 border-t border-slate-200 space-y-3">
                <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-900">
                  <BookOpen className="w-4 h-4 text-blue-600" />
                  <span>Статьи и исследования об объекте:</span>
                </div>
                <div className="space-y-2">
                  {STORIES.filter(s => s.relatedObjectSlugs?.includes(object.slug)).map((st) => (
                    <Link
                      key={st.id}
                      href={`/stories/${st.slug}`}
                      className="p-3 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-xl flex items-center justify-between group transition-colors"
                    >
                      <div className="min-w-0 flex-1 pr-2">
                        <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider block">
                          {st.categoryLabel}
                        </span>
                        <h4 className="text-xs sm:text-sm font-bold text-slate-950 group-hover:text-blue-600 transition-colors truncate">
                          {st.title}
                        </h4>
                      </div>
                      <span className="text-xs font-bold text-slate-900 group-hover:text-blue-600 flex items-center gap-1 shrink-0">
                        <span>Читать</span>
                        <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-0.5 transition-transform" />
                      </span>
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {/* Tags */}
            <div className="pt-3 border-t border-slate-100 flex flex-wrap gap-1.5">
              {object.tags.map((tag) => (
                <span
                  key={tag}
                  className="px-2 py-0.5 bg-slate-100 text-slate-700 rounded text-[11px] font-medium"
                >
                  #{tag}
                </span>
              ))}
            </div>
          </div>
        )}

        {/* TAB 2: Panorama 360 Player */}
        {activeTab === 'panorama' && hasPanorama(object) && (
          <PanoramaViewer key={object.id} panoramas={object.panoramas ?? []} heightClassName="h-72 sm:h-80" />
        )}

        {/* TAB 3: Audio Guide */}
        {activeTab === 'audio' && object.hasAudioGuide && (
          <div className="space-y-4">
            <div className="p-4 bg-slate-50 border border-slate-200 rounded-xl space-y-3">
              <div className="flex items-center justify-between">
                <div>
                  <h4 className="text-sm font-bold text-slate-900">
                    Аудиоэкскурсия к объекту
                  </h4>
                  <p className="text-xs text-slate-500">
                    Длительность: {object.audioDuration}
                  </p>
                </div>
                <button
                  onClick={() => setIsPlayingAudio(!isPlayingAudio)}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-semibold flex items-center gap-1.5 transition-colors"
                >
                  <Volume2 className="w-4 h-4" />
                  <span>{isPlayingAudio ? 'Пауза' : 'Слушать'}</span>
                </button>
              </div>

              <div className="p-2.5 bg-white border border-slate-200 rounded text-center text-xs text-slate-500 font-mono">
                [ 00:00 / {object.audioDuration} ]
              </div>
            </div>

            <p className="text-xs text-slate-600 leading-relaxed">
              Запись подготовлена историками и краеведами АНО «Центр сохранения архитектуры».
            </p>
          </div>
        )}

      </div>

      {/* Drawer Footer Actions */}
      <div className="p-4 border-t border-slate-200 bg-slate-50 flex items-center justify-between gap-3 shrink-0">
        <span className="text-xs font-mono text-slate-500">
          GPS: {object.coordinates.lat.toFixed(4)}, {object.coordinates.lng.toFixed(4)}
        </span>
        <button
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

    </div>
  );
}
