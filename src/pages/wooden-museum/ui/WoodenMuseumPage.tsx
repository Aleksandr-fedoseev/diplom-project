'use client';

import { useState } from 'react';
import Link from 'next/link';
import { ArrowLeft, ArrowRight, Building2, Heart } from 'lucide-react';
import { ARTIFACT_CATEGORIES, ArtifactCard, WOODEN_ARTIFACTS } from '@/entities/wooden-artifact';
import { cn } from '@/shared/lib';
import { AboutProject } from './AboutProject';

// Рубрики без экспонатов не показываем.
const CATEGORIES = [
  { id: 'all', label: 'Все предметы' },
  ...ARTIFACT_CATEGORIES.filter((c) => WOODEN_ARTIFACTS.some((a) => a.category === c.id)),
];

export default function WoodenMuseumPage() {
  const [category, setCategory] = useState<string>('all');
  const artifacts = WOODEN_ARTIFACTS.filter((a) => category === 'all' || a.category === category);

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      <div className="mb-6">
        <Link
          href="/projects"
          className="inline-flex items-center gap-1.5 text-xs font-bold text-slate-500 hover:text-slate-950 transition-colors uppercase tracking-wider"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Назад к списку проектов</span>
        </Link>
      </div>

      <div className="mb-12 space-y-3">
        <div className="inline-flex items-center gap-2 px-3 py-1 bg-slate-100 rounded-md text-xs font-bold uppercase tracking-wider text-slate-900">
          <Building2 className="w-3.5 h-3.5" />
          <span>Проект №2 • АНО «Центр сохранения архитектуры»</span>
        </div>
        <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-slate-950">Музей деревянной архитектуры</h1>
        <p className="text-base sm:text-lg text-slate-600 max-w-3xl leading-relaxed">
          Программа по спасению, научной консервации и созданию общедоступного каталога резного деревянного наследия Астрахани.
        </p>
      </div>

      <AboutProject />

      <section className="space-y-6 pt-4">
        <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4 border-b border-slate-200 pb-4">
          <div>
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Коллекция фонда</span>
            <h2 className="text-2xl sm:text-3xl font-black text-slate-950 tracking-tight mt-0.5">Спасенные элементы в коллекции</h2>
          </div>
          <div className="flex items-center gap-1.5 overflow-x-auto">
            {CATEGORIES.map((c) => (
              <button
                key={c.id}
                type="button"
                aria-pressed={category === c.id}
                onClick={() => setCategory(c.id)}
                className={cn(
                  'px-3 py-1.5 rounded-lg text-xs font-bold uppercase tracking-wider transition-colors whitespace-nowrap',
                  category === c.id ? 'bg-slate-950 text-white' : 'bg-slate-100 text-slate-700 hover:bg-slate-200',
                )}
              >
                {c.label}
              </button>
            ))}
          </div>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {artifacts.map((item) => (
            <ArtifactCard key={item.id} item={item} />
          ))}
        </div>
      </section>

      <section className="mt-16 p-8 sm:p-10 bg-slate-50 border border-slate-200 rounded-3xl flex flex-col sm:flex-row sm:items-center justify-between gap-6">
        <div className="space-y-2 max-w-xl">
          <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-500">
            <Heart className="w-4 h-4 text-rose-500" />
            <span>Участие жителей</span>
          </div>
          <h2 className="text-xl sm:text-2xl font-bold text-slate-950">Сохранился старый наличник, карниз или дверь?</h2>
          <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
            Не выбрасывайте историческую резьбу при демонтаже или замене окон. Передайте предмет в коллекцию музея — наши
            реставраторы сохранят его для будущих поколений.
          </p>
        </div>
        <Link
          href="/support?tab=wooden"
          className="inline-flex items-center justify-center gap-2 px-6 py-3.5 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs sm:text-sm font-bold uppercase tracking-wider transition-colors shrink-0"
        >
          <span>Передать в музей</span>
          <ArrowRight className="w-4 h-4" />
        </Link>
      </section>
    </div>
  );
}
