'use client';

import { useState } from 'react';
import { STORIES, STORY_CATEGORIES, StoryCard } from '@/entities/story';
import { cn } from '@/shared/lib';
import { Button, SearchInput } from '@/shared/ui';
import { filterStories } from '../model/filter';

const CATEGORIES = [{ id: 'all', label: 'Все' }, ...STORY_CATEGORIES];

export default function StoriesPage() {
  const [category, setCategory] = useState<string>('all');
  const [query, setQuery] = useState('');
  const filtered = filterStories(STORIES, category, query);
  const categoryLabel = CATEGORIES.find((c) => c.id === category)?.label;

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-14 w-full bg-white text-slate-900">
      <div className="mb-8">
        <div className="flex items-center gap-3">
          <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-slate-950">Истории</h1>
          <span className="px-3 py-1 bg-slate-100 text-slate-900 text-sm font-bold rounded-full font-mono">
            {STORIES.length}
          </span>
        </div>
        <p className="text-sm sm:text-base text-slate-600 mt-2 max-w-2xl">
          Краеведческая картотека: статьи об истории города, архитектуре, людях, архивных документах и разоблачение
          городских мифов.
        </p>
      </div>

      {/* Рубрики и поиск (ТЗ 5.1) */}
      <div className="space-y-4 mb-10 pb-6 border-b border-slate-200">
        <SearchInput className="w-full max-w-md" placeholder="Поиск по статьям и исследованиям..." value={query} onChange={setQuery} />
        <div className="flex items-center gap-1.5 overflow-x-auto pb-2 scrollbar-none">
          {CATEGORIES.map((c) => (
            <button
              key={c.id}
              type="button"
              aria-pressed={category === c.id}
              onClick={() => setCategory(c.id)}
              className={cn(
                'px-3.5 py-1.5 rounded-xl text-xs font-bold uppercase tracking-wider transition-all whitespace-nowrap',
                category === c.id ? 'bg-slate-950 text-white shadow-xs' : 'bg-slate-100 text-slate-700 hover:bg-slate-200',
              )}
            >
              {c.label}
            </button>
          ))}
        </div>
      </div>

      {filtered.length === 0 ? (
        <div className="p-12 text-center bg-slate-50 rounded-2xl border border-slate-200 space-y-4">
          <p className="text-sm sm:text-base text-slate-600">
            {query.trim() ? (
              <>По запросу «{query.trim()}» ничего не найдено.</>
            ) : (
              <>
                В рубрике <strong className="text-slate-950">«{categoryLabel}»</strong> пока нет опубликованных материалов.
              </>
            )}
          </p>
          <Button
            variant="secondary"
            onClick={() => {
              setCategory('all');
              setQuery('');
            }}
          >
            Показать все статьи
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filtered.map((story) => (
            <StoryCard key={story.id} story={story} />
          ))}
        </div>
      )}
    </div>
  );
}
