'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { Search, ArrowRight, BookOpen, Clock, Calendar } from 'lucide-react';
import { STORIES } from '@/data/stories';

// Категории строго по ТЗ 5.1
const CATEGORIES = [
  { id: 'all', label: 'Все' },
  { id: 'history', label: 'История города' },
  { id: 'architecture', label: 'Архитектура' },
  { id: 'people', label: 'Люди' },
  { id: 'streets', label: 'Улицы' },
  { id: 'houses', label: 'Дома' },
  { id: 'myths', label: 'Городские легенды' },
  { id: 'documents', label: 'Документы' },
  { id: 'photos', label: 'Фотографии' },
  { id: 'news', label: 'Новости' },
  { id: 'wooden', label: 'Деревянное зодчество' },
  { id: 'culture', label: 'Культура' },
  { id: 'research', label: 'Исследования' },
];

export default function StoriesPage() {
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');

  const filtered = STORIES.filter((item) => {
    const matchesCat = selectedCategory === 'all' || item.category === selectedCategory;
    const matchesSearch = searchQuery === '' || 
      item.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.lead.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesCat && matchesSearch;
  });

  const activeCategoryObj = CATEGORIES.find(c => c.id === selectedCategory);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-14 w-full bg-white text-slate-900">
      
      {/* 5. Заголовок раздела «Истории» */}
      <div className="mb-8">
        <div className="flex items-center gap-3">
          <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-slate-950">
            Истории
          </h1>
          <span className="px-3 py-1 bg-slate-100 text-slate-900 text-sm font-bold rounded-full font-mono">
            {STORIES.length}
          </span>
        </div>
        <p className="text-sm sm:text-base text-slate-600 mt-2 max-w-2xl">
          Краеведческая картотека: статьи об истории города, архитектуре, людях, архивных документах и разоблачение городских мифов.
        </p>
      </div>

      {/* 5.1. Категории и Поиск */}
      <div className="space-y-4 mb-10 pb-6 border-b border-slate-200">
        
        {/* Поисковая строка */}
        <div className="relative w-full max-w-md">
          <Search className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            type="text"
            placeholder="Поиск по статьям и исследованиям..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-slate-50 border border-slate-200 pl-10 pr-4 py-2.5 rounded-xl text-xs sm:text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950 focus:bg-white transition-all"
          />
        </div>

        {/* Список категорий (ТЗ 5.1) */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-2 scrollbar-none">
          {CATEGORIES.map((cat) => (
            <button
              key={cat.id}
              onClick={() => setSelectedCategory(cat.id)}
              className={`px-3.5 py-1.5 rounded-xl text-xs font-bold uppercase tracking-wider transition-all whitespace-nowrap ${
                selectedCategory === cat.id
                  ? 'bg-slate-950 text-white shadow-xs'
                  : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
              }`}
            >
              {cat.label}
            </button>
          ))}
        </div>

      </div>

      {/* Сетка статей по картотеке */}
      {filtered.length === 0 ? (
        <div className="p-12 text-center bg-slate-50 rounded-2xl border border-slate-200 space-y-4">
          <p className="text-sm sm:text-base text-slate-600">
            В категории <strong className="text-slate-950">«{activeCategoryObj?.label}»</strong> пока нет опубликованных материалов.
          </p>
          <button
            onClick={() => { setSelectedCategory('all'); setSearchQuery(''); }}
            className="px-5 py-2.5 bg-slate-950 text-white text-xs font-bold uppercase tracking-wider rounded-xl hover:bg-slate-800 transition-colors"
          >
            Показать все статьи
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filtered.map((story) => (
            <Link
              key={story.id}
              href={`/stories/${story.slug}`}
              className="group bg-white border border-slate-200 rounded-2xl overflow-hidden shadow-xs hover:shadow-md hover:border-slate-300 transition-all flex flex-col justify-between"
            >
              <div>
                <div className="h-52 bg-slate-100 overflow-hidden relative">
                  <img
                    src={story.coverImage}
                    alt={story.title}
                    className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                  />
                  <div className="absolute top-3 left-3">
                    <span className="px-2.5 py-1 bg-slate-950/85 backdrop-blur-md text-white text-[10px] font-bold uppercase tracking-wider rounded">
                      {story.categoryLabel}
                    </span>
                  </div>
                </div>

                <div className="p-5 space-y-2">
                  <div className="flex items-center gap-2 text-xs text-slate-400 font-medium">
                    <span>{story.publishedAt}</span>
                    <span>•</span>
                    <span>{story.readTime}</span>
                  </div>

                  <h3 className="text-base font-bold text-slate-950 group-hover:text-blue-600 transition-colors leading-snug">
                    {story.title}
                  </h3>

                  <p className="text-xs text-slate-600 line-clamp-3 leading-relaxed">
                    {story.lead}
                  </p>
                </div>
              </div>

              <div className="px-5 pb-5 pt-0 flex items-center justify-between text-xs font-semibold text-slate-900 group-hover:text-blue-600">
                <span className="text-[11px] text-slate-400 font-normal">{story.author.name}</span>
                <span className="flex items-center gap-1">
                  <span>Читать статью</span>
                  <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition-transform" />
                </span>
              </div>
            </Link>
          ))}
        </div>
      )}

    </div>
  );
}
