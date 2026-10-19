'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { Search, Filter, Eye, Volume2, ArrowRight, Building2, MapPin } from 'lucide-react';
import { HERITAGE_OBJECTS } from '@/data/heritage';

export default function ObjectsPage() {
  const [search, setSearch] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [selectedStatus, setSelectedStatus] = useState('all');

  const filtered = HERITAGE_OBJECTS.filter((item) => {
    const matchesCategory = selectedCategory === 'all' || item.category === selectedCategory;
    const matchesStatus = selectedStatus === 'all' || item.protectionStatus === selectedStatus;
    const matchesSearch = search === '' || 
      item.title.toLowerCase().includes(search.toLowerCase()) ||
      item.address.toLowerCase().includes(search.toLowerCase()) ||
      item.architecturalStyle.toLowerCase().includes(search.toLowerCase());
    return matchesCategory && matchesStatus && matchesSearch;
  });

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-14 w-full">
      
      {/* Page Header */}
      <div className="mb-8">
        <div className="text-xs font-bold uppercase tracking-wider text-blue-600 mb-1">
          Памятники Астрахани
        </div>
        <h1 className="text-3xl sm:text-4xl font-bold text-slate-900 tracking-tight">
          Исторические здания и архитектурные памятники
        </h1>
        <p className="text-base text-slate-600 mt-2 max-w-2xl">
          Собрание старинных купеческих усадеб, деревянных теремов, кремлевских построек и утраченных храмов Астраханской области.
        </p>
      </div>

      {/* Filter & Search Bar */}
      <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs mb-8 space-y-4">
        <div className="flex flex-col md:flex-row gap-3">
          <div className="relative flex-1">
            <Search className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              placeholder="Поиск по названию, адресу, архитектору..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full bg-slate-50 border border-slate-200 pl-10 pr-4 py-2.5 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
            />
          </div>

          <div className="flex items-center gap-2">
            <select
              value={selectedCategory}
              onChange={(e) => setSelectedCategory(e.target.value)}
              className="bg-slate-50 border border-slate-200 px-3 py-2.5 rounded-lg text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-blue-600"
            >
              <option value="all">Все категории</option>
              <option value="wooden">Деревянное зодчество</option>
              <option value="manor">Купеческие усадьбы</option>
              <option value="temple">Кремль и храмы</option>
              <option value="modern">Модерн</option>
              <option value="lost">Утраченное</option>
            </select>

            <select
              value={selectedStatus}
              onChange={(e) => setSelectedStatus(e.target.value)}
              className="bg-slate-50 border border-slate-200 px-3 py-2.5 rounded-lg text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-blue-600"
            >
              <option value="all">Любой статус охраны</option>
              <option value="federal">Федеральное значение</option>
              <option value="regional">Региональное значение</option>
              <option value="lost">Утраченный памятник</option>
            </select>
          </div>
        </div>
      </div>

      {/* Grid of Objects */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {filtered.map((item) => (
          <Link
            key={item.id}
            href="/map"
            className="bg-white rounded-xl border border-slate-200 overflow-hidden shadow-xs hover:shadow-md hover:border-slate-300 transition-all flex flex-col justify-between group"
          >
            <div>
              {item.coverImage ? (
                <div className="h-48 w-full bg-slate-100 overflow-hidden relative">
                  <img
                    src={item.coverImage}
                    alt={item.title}
                    className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                  />
                  <div className="absolute top-3 right-3 flex gap-1.5">
                    {item.hasPanorama360 && (
                      <span className="bg-slate-900/85 backdrop-blur-md text-white text-[10px] font-bold px-2 py-0.5 rounded shadow flex items-center gap-1">
                        <Eye className="w-3 h-3 text-blue-400" />
                        <span>360°</span>
                      </span>
                    )}
                  </div>
                </div>
              ) : (
                <div className="h-32 w-full bg-slate-50 flex items-center justify-center text-slate-400 border-b border-slate-100">
                  <Building2 className="w-8 h-8" />
                </div>
              )}

              <div className="p-5">
                <div className="flex items-center justify-between gap-2 mb-2">
                  <span className="text-[11px] font-semibold text-blue-600 uppercase">
                    {item.yearBuilt} г. • {item.architecturalStyle}
                  </span>
                </div>

                <h3 className="text-base font-bold text-slate-900 group-hover:text-blue-600 transition-colors mb-1.5 leading-snug">
                  {item.title}
                </h3>

                <p className="text-xs text-slate-500 mb-2.5 flex items-center gap-1">
                  <MapPin className="w-3.5 h-3.5 shrink-0 text-slate-400" />
                  <span className="truncate">{item.address}</span>
                </p>

                <p className="text-xs text-slate-600 line-clamp-2 leading-relaxed">
                  {item.shortDescription}
                </p>
              </div>
            </div>

            <div className="px-5 pb-5 pt-0">
              <div className="pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
                <span className={`font-semibold ${
                  item.protectionStatus === 'federal' ? 'text-emerald-700' :
                  item.protectionStatus === 'lost' ? 'text-rose-700' : 'text-slate-600'
                }`}>
                  {item.protectionStatus === 'federal' && 'Федеральный ОКН'}
                  {item.protectionStatus === 'regional' && 'Региональный ОКН'}
                  {item.protectionStatus === 'lost' && 'Утрачено'}
                </span>

                <span className="text-blue-600 font-semibold group-hover:translate-x-0.5 transition-transform">
                  Смотреть на карте →
                </span>
              </div>
            </div>
          </Link>
        ))}
      </div>

    </div>
  );
}
