'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { 
  Building2, 
  ArrowLeft, 
  HelpCircle, 
  Flame, 
  Layers, 
  Warehouse, 
  Sparkles, 
  Heart,
  Calendar,
  CheckCircle2,
  Clock,
  ArrowRight
} from 'lucide-react';
import { WOODEN_ARTIFACTS } from '@/data/woodenMuseum';

export default function WoodenMuseumPage() {
  const [activeCategory, setActiveCategory] = useState<string>('all');

  const categories = [
    { id: 'all', label: 'Все предметы' },
    { id: 'platband', label: 'Наличники' },
    { id: 'door', label: 'Двери' },
    { id: 'cornice', label: 'Карнизы' },
  ];

  const filteredArtifacts = activeCategory === 'all'
    ? WOODEN_ARTIFACTS
    : WOODEN_ARTIFACTS.filter(a => a.category === activeCategory);

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      
      {/* Навигация назад */}
      <div className="mb-6">
        <Link
          href="/projects"
          className="inline-flex items-center gap-1.5 text-xs font-bold text-slate-500 hover:text-slate-950 transition-colors uppercase tracking-wider"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Назад к списку проектов</span>
        </Link>
      </div>

      {/* 8. Заголовок Проекта №2 */}
      <div className="mb-12 space-y-3">
        <div className="inline-flex items-center gap-2 px-3 py-1 bg-slate-100 rounded-md text-xs font-bold uppercase tracking-wider text-slate-900">
          <Building2 className="w-3.5 h-3.5" />
          <span>Проект №2 • АНО «Центр сохранения архитектуры»</span>
        </div>
        <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-slate-950">
          Музей деревянной архитектуры
        </h1>
        <p className="text-base sm:text-lg text-slate-600 max-w-3xl leading-relaxed">
          Программа по спасению, научной консервации и созданию общедоступного каталога резного деревянного наследия Астрахани.
        </p>
      </div>

      {/* ========================================================================= */}
      {/* Ответы на ключевые вопросы по ТЗ 8                                        */}
      {/* ========================================================================= */}
      <div className="space-y-6 mb-16">
        <h2 className="text-xl sm:text-2xl font-black text-slate-950 tracking-tight">
          О проекте сохранения деревянного зодчества
        </h2>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          
          {/* 1. Зачем создан проект */}
          <div className="p-6 sm:p-7 bg-slate-50 border border-slate-200 rounded-2xl space-y-3 flex flex-col justify-between">
            <div className="space-y-2">
              <div className="flex items-center gap-2.5 text-xs font-bold uppercase tracking-wider text-slate-900">
                <HelpCircle className="w-4 h-4 text-blue-600" />
                <span>Зачем создан проект</span>
              </div>
              <h3 className="text-base sm:text-lg font-bold text-slate-950">
                Спасение подлинных деталей от безвозвратной утраты
              </h3>
              <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
                Деревянные дома исторических слобод города ветшают, попадают под программы расселения и сносятся. Проект создан для того, чтобы ценнейшие образцы резного декора — наличники, филенчатые двери, карнизы и подзоры — не уходили на свалку, а бережно демонтировались, сохранялись и становились основой будущего музея.
              </p>
            </div>
          </div>

          {/* 2. Почему необходимо сохранять */}
          <div className="p-6 sm:p-7 bg-slate-50 border border-slate-200 rounded-2xl space-y-3 flex flex-col justify-between">
            <div className="space-y-2">
              <div className="flex items-center gap-2.5 text-xs font-bold uppercase tracking-wider text-slate-900">
                <Flame className="w-4 h-4 text-amber-600" />
                <span>Почему необходимо сохранять</span>
              </div>
              <h3 className="text-base sm:text-lg font-bold text-slate-950">
                Уникальность южнорусской плотницкой традиции
              </h3>
              <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
                В Астрахани нет местных строевых лесов: вековые сосны и лиственницы доставлялись за тысячи километров по Волге на гигантских барках-белянах. Мастера создали неповторимый сплав северной корабельной глухой резьбы с солярными южными знаками и восточным растительным орнаментом, не имеющий аналогов в России.
              </p>
            </div>
          </div>

          {/* 3. Как проходит сбор */}
          <div className="p-6 sm:p-7 bg-slate-50 border border-slate-200 rounded-2xl space-y-3 flex flex-col justify-between">
            <div className="space-y-2">
              <div className="flex items-center gap-2.5 text-xs font-bold uppercase tracking-wider text-slate-900">
                <Layers className="w-4 h-4 text-emerald-600" />
                <span>Как проходит сбор</span>
              </div>
              <h3 className="text-base sm:text-lg font-bold text-slate-950">
                Экспедиционный мониторинг и бережный демонтаж
              </h3>
              <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
                Команда Центра ведет учет аварийного деревянного фонда. При подготовке дома к демонтажу волонтеры и реставраторы согласуют вывоз элементов, деликатно снимают детали без раскола древесины, проводят фотофиксацию, научные обмеры и составляют исторический паспорт каждого предмета. Также элементы передают неравнодушные жители.
              </p>
            </div>
          </div>

          {/* 4. Где хранятся элементы */}
          <div className="p-6 sm:p-7 bg-slate-50 border border-slate-200 rounded-2xl space-y-3 flex flex-col justify-between">
            <div className="space-y-2">
              <div className="flex items-center gap-2.5 text-xs font-bold uppercase tracking-wider text-slate-900">
                <Warehouse className="w-4 h-4 text-purple-600" />
                <span>Где хранятся элементы</span>
              </div>
              <h3 className="text-base sm:text-lg font-bold text-slate-950">
                Специализированное фондохранилище
              </h3>
              <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
                Все собранные артефакты размещаются в сухом вентилируемом фондохранилище мастерской Центра. Древесина проходит очистку от пыли и копоти, антисептическую противогрибковую обработку и стабилизирующую консервацию для предотвращения растрескивания и защиты от влаги.
              </p>
            </div>
          </div>

        </div>

        {/* 5. Планы на будущее */}
        <div className="p-6 sm:p-8 bg-slate-950 text-white rounded-2xl space-y-3">
          <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-300">
            <Sparkles className="w-4 h-4 text-amber-400" />
            <span>Планы на будущее</span>
          </div>
          <h3 className="text-xl sm:text-2xl font-bold text-white tracking-tight">
            Открытая музейная экспозиция и ремесленная школа
          </h3>
          <p className="text-xs sm:text-sm text-slate-300 leading-relaxed max-w-3xl">
            Главная цель проекта — открыть постоянное городское музейное пространство под открытым небом, где каждый сможет увидеть шедевры деревянной резьбы вблизи. В планах также запуск открытой плотницкой мастерской по обучению традиционной резьбе и создание полной 3D-модели каждого спасенного наличника.
          </p>
        </div>

      </div>

      {/* ========================================================================= */}
      {/* 6. Какие предметы находятся в коллекции / 3. Какие элементы удалось спасти */}
      {/* ========================================================================= */}
      <div className="space-y-6 pt-4">
        <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4 border-b border-slate-200 pb-4">
          <div>
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
              Коллекция фонда
            </span>
            <h2 className="text-2xl sm:text-3xl font-black text-slate-950 tracking-tight mt-0.5">
              Спасенные элементы в коллекции
            </h2>
          </div>

          {/* Фильтры по типу предметов */}
          <div className="flex items-center gap-1.5 overflow-x-auto">
            {categories.map((cat) => (
              <button
                key={cat.id}
                onClick={() => setActiveCategory(cat.id)}
                className={`px-3 py-1.5 rounded-lg text-xs font-bold uppercase tracking-wider transition-colors whitespace-nowrap ${
                  activeCategory === cat.id
                    ? 'bg-slate-950 text-white'
                    : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
                }`}
              >
                {cat.label}
              </button>
            ))}
          </div>
        </div>

        {/* Сетка предметов */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredArtifacts.map((item) => (
            <div
              key={item.id}
              className="bg-white border border-slate-200 rounded-2xl overflow-hidden shadow-xs hover:shadow-md transition-all flex flex-col justify-between group"
            >
              <div>
                <div className="h-56 bg-slate-100 relative overflow-hidden">
                  <img
                    src={item.coverImage}
                    alt={item.title}
                    className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                  />
                  <div className="absolute top-3 left-3">
                    <span className="px-2.5 py-1 bg-slate-950/85 backdrop-blur-md text-white text-[10px] font-bold uppercase tracking-wider rounded">
                      {item.categoryLabel}
                    </span>
                  </div>
                  <div className="absolute bottom-3 right-3">
                    <span className="px-2.5 py-1 bg-white/90 backdrop-blur-md text-slate-900 text-[10px] font-bold rounded shadow-xs">
                      {item.statusLabel}
                    </span>
                  </div>
                </div>

                <div className="p-5 space-y-2.5">
                  <h3 className="text-base font-bold text-slate-950 leading-snug">
                    {item.title}
                  </h3>

                  <div className="text-xs text-slate-500 space-y-1">
                    <div className="font-medium text-slate-700">
                      {item.originalAddress}
                    </div>
                    <div className="text-[11px] text-slate-400">
                      Период создания: {item.yearBuilt}
                    </div>
                  </div>

                  <p className="text-xs text-slate-600 leading-relaxed pt-2 border-t border-slate-100">
                    {item.description}
                  </p>
                </div>
              </div>

              <div className="px-5 pb-5 pt-0 text-[11px] text-slate-400 font-mono flex items-center justify-between">
                <span>Год спасения:</span>
                <span className="font-bold text-slate-700">{item.acquisitionYear}</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* ========================================================================= */}
      {/* Призыв к участию (Передать наличник в музей)                               */}
      {/* ========================================================================= */}
      <div className="mt-16 p-8 sm:p-10 bg-slate-50 border border-slate-200 rounded-3xl flex flex-col sm:flex-row sm:items-center justify-between gap-6">
        <div className="space-y-2 max-w-xl">
          <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-500">
            <Heart className="w-4 h-4 text-rose-500" />
            <span>Участие жителей</span>
          </div>
          <h3 className="text-xl sm:text-2xl font-bold text-slate-950">
            Сохранился старый наличник, карниз или дверь?
          </h3>
          <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
            Не выбрасывайте историческую резьбу при демонтаже или замене окон. Передайте предмет в коллекцию музея — наши реставраторы сохранят его для будущих поколений.
          </p>
        </div>
        <Link
          href="/support#wooden"
          className="inline-flex items-center justify-center gap-2 px-6 py-3.5 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs sm:text-sm font-bold uppercase tracking-wider transition-colors shrink-0"
        >
          <span>Передать в музей</span>
          <ArrowRight className="w-4 h-4" />
        </Link>
      </div>

    </div>
  );
}
