'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import dynamic from 'next/dynamic';
import { 
  Search, 
  Eye, 
  X, 
  MapPin,
  Building2,
  Layers,
  Info
} from 'lucide-react';
import { HERITAGE_OBJECTS } from '@/data/heritage';
import { HeritageObject } from '@/types/heritage';
import ObjectDetailDrawer from '@/components/map/ObjectDetailDrawer';

import { Suspense, useEffect } from 'react';
import { useSearchParams } from 'next/navigation';

// Dynamically import Leaflet map component for client-side rendering
const AstrakhanMap = dynamic(() => import('@/components/map/AstrakhanMap'), {
  ssr: false,
  loading: () => (
    <div className="w-full h-full bg-slate-100 flex flex-col items-center justify-center text-slate-500 gap-3">
      <div className="w-8 h-8 border-3 border-blue-600 border-t-transparent rounded-full animate-spin" />
      <span className="text-xs font-semibold">Загрузка OpenStreetMap Астраханской области...</span>
    </div>
  ),
});

function MapContent() {
  const searchParams = useSearchParams();
  const objectSlug = searchParams.get('object');

  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [only360, setOnly360] = useState(false);
  const [activeObject, setActiveObject] = useState<HeritageObject | null>(null);

  // Auto-select object if passed in URL: /map?object=dom-tetushinova
  useEffect(() => {
    if (objectSlug) {
      const target = HERITAGE_OBJECTS.find(
        (o) => o.slug === objectSlug || o.id === objectSlug
      );
      if (target) {
        setActiveObject(target);
      }
    }
  }, [objectSlug]);

  const filtered = HERITAGE_OBJECTS.filter((item) => {
    const matchesCategory = selectedCategory === 'all' || item.category === selectedCategory;
    const matches360 = !only360 || item.hasPanorama360;
    const matchesSearch = searchQuery === '' || 
      item.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.address.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.architecturalStyle.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesCategory && matches360 && matchesSearch;
  });

  return (
    <div className="h-[calc(100vh-4rem)] sm:h-[calc(100vh-5rem)] flex flex-col bg-slate-100 overflow-hidden relative">
      
      {/* Top Filter Bar */}
      <div className="bg-white border-b border-slate-200 px-4 py-3 z-20 flex flex-wrap items-center justify-between gap-3 shadow-xs">
        
        {/* Search */}
        <div className="relative flex-1 min-w-[220px] max-w-md">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            type="text"
            placeholder="Поиск по памятникам Астрахани..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-slate-50 border border-slate-200 pl-9 pr-3 py-2 rounded-lg text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
          />
          {searchQuery && (
            <button 
              onClick={() => setSearchQuery('')}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>

        {/* Category Filters */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 sm:pb-0 scrollbar-none">
          {[
            { id: 'all', label: 'Все объекты' },
            { id: 'wooden', label: 'Деревянное зодчество' },
            { id: 'manor', label: 'Купеческие усадьбы' },
            { id: 'temple', label: 'Кремль и храмы' },
            { id: 'modern', label: 'Модерн' },
            { id: 'lost', label: 'Утраченное' },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setSelectedCategory(tab.id)}
              className={`px-3 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors ${
                selectedCategory === tab.id
                  ? 'bg-slate-900 text-white'
                  : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
              }`}
            >
              {tab.label}
            </button>
          ))}

          <button
            onClick={() => setOnly360(!only360)}
            className={`px-3 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap flex items-center gap-1.5 transition-colors ${
              only360
                ? 'bg-blue-600 text-white'
                : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
            }`}
          >
            <Eye className="w-3.5 h-3.5" />
            <span>Только с 360°</span>
          </button>
        </div>

        {/* Region Lock Badge */}
        <div className="text-xs font-medium text-slate-500 hidden xl:flex items-center gap-1.5 bg-slate-100 px-2.5 py-1.5 rounded-lg border border-slate-200">
          <span className="w-2 h-2 rounded-full bg-emerald-500" />
          <span>Карта OpenStreetMap (Астраханская область)</span>
        </div>
      </div>

      {/* Main Split Layout: Left List + Right Real OSM Map */}
      <div className="flex-1 flex overflow-hidden relative">
        
        {/* Left Sidebar (Desktop 360px, Mobile toggleable) */}
        <div className="w-full md:w-[340px] lg:w-[380px] bg-white border-r border-slate-200 flex flex-col shrink-0 overflow-hidden z-10">
          
          <div className="p-3 border-b border-slate-100 bg-slate-50 text-xs font-medium text-slate-600 flex justify-between items-center">
            <span>Объектов: {filtered.length}</span>
            <span className="text-slate-400">Кликните для просмотра</span>
          </div>

          {/* Cards Scrollable Container */}
          <div className="flex-1 overflow-y-auto divide-y divide-slate-100">
            {filtered.length === 0 ? (
              <div className="p-8 text-center text-slate-500 text-sm">
                По заданному фильтру объектов не найдено.
              </div>
            ) : (
              filtered.map((item) => {
                const isSelected = activeObject?.id === item.id;
                return (
                  <div
                    key={item.id}
                    onClick={() => setActiveObject(item)}
                    className={`p-3.5 cursor-pointer transition-colors flex gap-3 ${
                      isSelected ? 'bg-blue-50/80 border-l-4 border-blue-600' : 'hover:bg-slate-50'
                    }`}
                  >
                    {item.coverImage ? (
                      <div className="w-16 h-16 rounded-lg overflow-hidden bg-slate-100 shrink-0 border border-slate-200 relative">
                        <img
                          src={item.coverImage}
                          alt={item.title}
                          className="w-full h-full object-cover"
                        />
                      </div>
                    ) : (
                      <div className="w-16 h-16 rounded-lg bg-slate-100 flex items-center justify-center text-slate-400 shrink-0 border border-slate-200">
                        <Building2 className="w-6 h-6" />
                      </div>
                    )}

                    <div className="flex-1 min-w-0 flex flex-col justify-between">
                      <div>
                        <div className="flex items-center justify-between gap-1">
                          <span className="text-[10px] font-semibold text-blue-600 uppercase truncate">
                            {item.yearBuilt} г. • {item.architecturalStyle}
                          </span>
                          {item.hasPanorama360 && (
                            <span className="bg-blue-600 text-white text-[8px] font-bold px-1 rounded shrink-0">
                              360°
                            </span>
                          )}
                        </div>

                        <h3 className="text-xs sm:text-sm font-bold text-slate-900 leading-snug truncate mt-0.5">
                          {item.title}
                        </h3>

                        <p className="text-[11px] text-slate-500 truncate flex items-center gap-1 mt-0.5">
                          <MapPin className="w-3 h-3 text-slate-400 shrink-0" />
                          <span>{item.address}</span>
                        </p>
                      </div>

                      <div className="flex items-center justify-between pt-1 text-[10px]">
                        <span className={`font-semibold px-1.5 py-0.5 rounded ${
                          item.protectionStatus === 'federal'
                            ? 'bg-emerald-50 text-emerald-800'
                            : item.protectionStatus === 'lost'
                            ? 'bg-rose-50 text-rose-800'
                            : 'bg-slate-100 text-slate-700'
                        }`}>
                          {item.protectionStatus === 'federal' && 'Фед. ОКН'}
                          {item.protectionStatus === 'regional' && 'Регион. ОКН'}
                          {item.protectionStatus === 'lost' && 'Утрачено'}
                        </span>

                        <span className="font-semibold text-blue-600">
                          Подробнее →
                        </span>
                      </div>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>

        {/* Right Real OpenStreetMap Canvas */}
        <div className="flex-1 relative overflow-hidden bg-slate-200">
          <AstrakhanMap
            objects={filtered}
            activeObject={activeObject}
            onSelectObject={(obj) => setActiveObject(obj)}
          />
        </div>

        {/* In-Map Object Detail Drawer (Slide-over without page navigation) */}
        <ObjectDetailDrawer
          object={activeObject}
          onClose={() => setActiveObject(null)}
        />

      </div>

    </div>
  );
}

export default function MapPage() {
  return (
    <Suspense fallback={
      <div className="h-[calc(100vh-4rem)] sm:h-[calc(100vh-5rem)] flex items-center justify-center bg-slate-100">
        <div className="w-8 h-8 border-3 border-blue-600 border-t-transparent rounded-full animate-spin" />
      </div>
    }>
      <MapContent />
    </Suspense>
  );
}
