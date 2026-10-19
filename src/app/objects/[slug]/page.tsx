'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { notFound, useParams } from 'next/navigation';
import { 
  Eye, 
  Volume2, 
  MapPin, 
  ShieldCheck, 
  Share2, 
  ArrowLeft,
  RotateCw,
  Maximize2,
  FileText,
  Compass
} from 'lucide-react';
import { HERITAGE_OBJECTS } from '@/data/heritage';

export default function ObjectDetailPage() {
  const params = useParams();
  const slug = params?.slug as string;

  const object = HERITAGE_OBJECTS.find((item) => item.slug === slug) || HERITAGE_OBJECTS[0];
  const [activeTab, setActiveTab] = useState<'passport' | 'panorama' | 'audio'>('passport');
  const [isPlayingAudio, setIsPlayingAudio] = useState(false);

  if (!object) {
    return notFound();
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12 w-full">
      
      {/* Top Breadcrumbs & Back */}
      <div className="flex items-center justify-between gap-4 mb-6">
        <Link
          href="/objects"
          className="inline-flex items-center gap-1.5 text-sm font-semibold text-slate-600 hover:text-slate-900 transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Вернуться в каталог</span>
        </Link>

        <div className="flex items-center gap-2">
          <button 
            onClick={() => alert('Ссылка на паспорт скопирована!')}
            className="p-2 bg-white rounded-lg border border-slate-200 text-slate-600 hover:text-slate-900 hover:bg-slate-50 transition-colors text-xs flex items-center gap-1.5"
            title="Поделиться"
          >
            <Share2 className="w-4 h-4" />
            <span className="hidden sm:inline">Поделиться</span>
          </button>
        </div>
      </div>

      {/* Main Object Header Card */}
      <div className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-xs mb-8">
        <div className="flex flex-wrap items-center gap-2 mb-3">
          <span className={`text-xs font-bold uppercase tracking-wider px-2.5 py-1 rounded ${
            object.protectionStatus === 'federal'
              ? 'bg-emerald-50 text-emerald-800 border border-emerald-200'
              : object.protectionStatus === 'lost'
              ? 'bg-rose-50 text-rose-800 border border-rose-200'
              : 'bg-slate-100 text-slate-800 border border-slate-200'
          }`}>
            {object.protectionStatus === 'federal' && 'Памятник федерального значения'}
            {object.protectionStatus === 'regional' && 'Региональное наследие'}
            {object.protectionStatus === 'lost' && 'Утраченный объект'}
          </span>

          {object.registryNumber && (
            <span className="text-xs font-mono font-medium text-slate-500 bg-slate-100 px-2.5 py-1 rounded">
              № {object.registryNumber}
            </span>
          )}
        </div>

        <h1 className="text-3xl sm:text-4xl font-bold text-slate-900 tracking-tight mb-2">
          {object.title}
        </h1>

        <p className="text-sm sm:text-base text-slate-600 flex items-center gap-1.5 mb-6">
          <MapPin className="w-4 h-4 text-slate-400 shrink-0" />
          <span>{object.address}</span>
        </p>

        {/* Media / Interaction Tabs */}
        <div className="flex items-center gap-2 border-b border-slate-200 overflow-x-auto pb-0.5">
          <button
            onClick={() => setActiveTab('passport')}
            className={`px-4 py-3 text-sm font-semibold border-b-2 whitespace-nowrap transition-colors flex items-center gap-2 ${
              activeTab === 'passport'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <FileText className="w-4 h-4" />
            <span>Паспорт и историческая справка</span>
          </button>

          {object.hasPanorama360 && (
            <button
              onClick={() => setActiveTab('panorama')}
              className={`px-4 py-3 text-sm font-semibold border-b-2 whitespace-nowrap transition-colors flex items-center gap-2 ${
                activeTab === 'panorama'
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-slate-600 hover:text-slate-900'
              }`}
            >
              <Eye className="w-4 h-4" />
              <span>Сферическая панорама 360°</span>
              <span className="text-[10px] bg-blue-100 text-blue-800 px-1.5 py-0.5 rounded font-bold">
                VR
              </span>
            </button>
          )}

          {object.hasAudioGuide && (
            <button
              onClick={() => setActiveTab('audio')}
              className={`px-4 py-3 text-sm font-semibold border-b-2 whitespace-nowrap transition-colors flex items-center gap-2 ${
                activeTab === 'audio'
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-slate-600 hover:text-slate-900'
              }`}
            >
              <Volume2 className="w-4 h-4" />
              <span>Аудиогид ({object.audioDuration})</span>
            </button>
          )}
        </div>
      </div>

      {/* Tab Content Display */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        
        {/* Main Content Area (8 cols) */}
        <div className="lg:col-span-8 space-y-6">
          
          {/* TAB 1: Passport and History */}
          {activeTab === 'passport' && (
            <div className="bg-white rounded-2xl p-6 sm:p-8 border border-slate-200 shadow-xs space-y-6">
              <div>
                <h3 className="text-xl font-bold text-slate-900 mb-3">
                  Архитектурно-историческое описание
                </h3>
                <p className="text-base text-slate-700 leading-relaxed">
                  {object.fullHistory}
                </p>
              </div>

              <div className="pt-4 border-t border-slate-100">
                <h4 className="text-base font-bold text-slate-900 mb-2">
                  Краткая характеристика
                </h4>
                <p className="text-sm text-slate-600 leading-relaxed">
                  {object.shortDescription}
                </p>
              </div>

              {object.tags && (
                <div className="pt-4 border-t border-slate-100 flex flex-wrap gap-2">
                  {object.tags.map((tag) => (
                    <span
                      key={tag}
                      className="px-2.5 py-1 bg-slate-100 text-slate-700 rounded text-xs font-medium"
                    >
                      #{tag}
                    </span>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* TAB 2: 360° Panorama Viewer */}
          {activeTab === 'panorama' && object.hasPanorama360 && (
            <div className="space-y-4">
              <div className="relative h-[480px] sm:h-[540px] rounded-2xl overflow-hidden bg-slate-950 border border-slate-800 shadow-xl flex flex-col justify-between p-6">
                
                <div className="flex items-center justify-between z-10">
                  <div className="bg-slate-900/90 text-white text-xs font-semibold px-3 py-1.5 rounded-lg backdrop-blur-md">
                    Photo Sphere Viewer 360° • WebGL
                  </div>
                  <div className="flex gap-2">
                    <button className="p-2 bg-slate-900/90 text-white rounded-lg hover:bg-slate-800 transition-colors">
                      <RotateCw className="w-4 h-4" />
                    </button>
                    <button className="p-2 bg-slate-900/90 text-white rounded-lg hover:bg-slate-800 transition-colors">
                      <Maximize2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>

                <div className="text-center text-slate-300 text-sm bg-slate-900/80 p-6 rounded-xl self-center max-w-md backdrop-blur-md border border-slate-700">
                  <Eye className="w-8 h-8 text-blue-400 mx-auto mb-2 animate-pulse" />
                  <p className="font-semibold text-white">Интерактивный плеер сферической панорамы</p>
                  <p className="text-xs text-slate-400 mt-1">
                    Эквидистантная проекция 8K с хотспотами интерьеров и фасада памятника.
                  </p>
                </div>

                <div className="text-center text-white/80 text-xs bg-slate-900/70 py-1.5 px-4 rounded-lg self-center backdrop-blur-md">
                  Управление: ЛКМ/Сенсор для вращения обзора
                </div>
              </div>
            </div>
          )}

          {/* TAB 3: Audio Guide */}
          {activeTab === 'audio' && object.hasAudioGuide && (
            <div className="bg-white rounded-2xl p-6 sm:p-8 border border-slate-200 shadow-xs space-y-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-xl font-bold text-slate-900">
                    Аудиоэкскурсия по объекту
                  </h3>
                  <p className="text-xs text-slate-500 mt-0.5">
                    Длительность: {object.audioDuration} • АНО «Центр сохранения архитектуры»
                  </p>
                </div>
                <button
                  onClick={() => setIsPlayingAudio(!isPlayingAudio)}
                  className="px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-semibold flex items-center gap-2 transition-colors"
                >
                  <Volume2 className="w-4 h-4" />
                  <span>{isPlayingAudio ? 'Пауза' : 'Слушать'}</span>
                </button>
              </div>

              <div className="p-4 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-600">
                [ Аудиоплеер синхронизирован с ключевыми точками фасада ]
              </div>
            </div>
          )}

        </div>

        {/* Right Metadata Sidebar (4 cols) */}
        <div className="lg:col-span-4 space-y-6">
          <div className="bg-white rounded-2xl p-6 border border-slate-200 shadow-xs space-y-5">
            <h3 className="text-base font-bold text-slate-900 border-b border-slate-100 pb-3">
              Паспортные реквизиты
            </h3>

            <div className="space-y-3.5 text-sm">
              <div>
                <div className="text-xs text-slate-400">Период постройки</div>
                <div className="font-semibold text-slate-900">{object.yearBuilt} г.</div>
              </div>

              <div>
                <div className="text-xs text-slate-400">Архитектурный стиль</div>
                <div className="font-semibold text-slate-900">{object.architecturalStyle}</div>
              </div>

              {object.architect && (
                <div>
                  <div className="text-xs text-slate-400">Архитектор / Зодчий</div>
                  <div className="font-semibold text-slate-900">{object.architect}</div>
                </div>
              )}

              <div>
                <div className="text-xs text-slate-400">Адрес объекта</div>
                <div className="font-semibold text-slate-900">{object.address}</div>
              </div>

              <div>
                <div className="text-xs text-slate-400">Географические координаты (WGS 84)</div>
                <div className="font-mono text-xs text-slate-700">
                  {object.coordinates.lat.toFixed(4)}° N, {object.coordinates.lng.toFixed(4)}° E
                </div>
              </div>
            </div>

            <div className="pt-3 border-t border-slate-100">
              <Link
                href="/map"
                className="w-full inline-flex items-center justify-center gap-2 py-2.5 px-4 bg-slate-900 hover:bg-blue-600 text-white rounded-lg text-xs font-semibold transition-colors"
              >
                <MapPin className="w-4 h-4 text-blue-400" />
                <span>Показать на интерактивной карте</span>
              </Link>
            </div>
          </div>
        </div>

      </div>

    </div>
  );
}
