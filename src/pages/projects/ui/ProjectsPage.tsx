import React from 'react';
import Link from 'next/link';
import { Compass, Building2, ArrowRight, ShieldCheck, Hammer } from 'lucide-react';

export default function ProjectsPage() {
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      
      {/* Header */}
      <div className="mb-12">
        <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
          Инициативы АНО «Центр сохранения архитектуры»
        </span>
        <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-slate-950 mt-1">
          Собственные проекты
        </h1>
        <p className="text-base sm:text-lg text-slate-600 mt-2 max-w-2xl">
          Практическая работа по оцифровке, архивации и физическому сохранению культурного наследия Астраханской области.
        </p>
      </div>

      {/* Projects Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        
        {/* Project 1: Интерактивная карта */}
        <div className="bg-slate-950 text-white rounded-3xl p-8 sm:p-10 flex flex-col justify-between relative overflow-hidden group shadow-lg">
          <div 
            className="absolute inset-0 bg-cover bg-center opacity-30 group-hover:scale-105 transition-transform duration-500"
            style={{ backgroundImage: 'url(/heritage/birzha/cover.jpg)' }}
          />
          <div className="absolute inset-0 bg-gradient-to-t from-slate-950 via-slate-950/70 to-transparent" />

          <div className="relative z-10 space-y-4">
            <span className="px-3 py-1 bg-white/20 backdrop-blur-md rounded text-[10px] font-bold uppercase tracking-wider text-white">
              Флагманский геосервис
            </span>
            <h2 className="text-2xl sm:text-3xl font-black tracking-tight text-white">
              Интерактивная карта культурного наследия
            </h2>
            <p className="text-sm sm:text-base text-slate-300 leading-relaxed">
              Интерактивная карта памятников и исторических зданий города. Удобный поиск по улицам, архивные фотографии, описания архитектурных стилей и просмотр истории каждого дома прямо на карте Астрахани.
            </p>
          </div>

          <div className="relative z-10 pt-8">
            <Link
              href="/map"
              className="inline-flex items-center gap-2 px-6 py-3.5 bg-white hover:bg-slate-100 text-slate-950 rounded-xl text-xs font-bold uppercase tracking-wider transition-colors shadow-sm"
            >
              <Compass className="w-4 h-4" />
              <span>Перейти к интерактивной карте</span>
            </Link>
          </div>
        </div>

        {/* Project 2: Музей деревянной архитектуры */}
        <div className="bg-slate-50 border border-slate-200 rounded-3xl p-8 sm:p-10 flex flex-col justify-between relative overflow-hidden group">
          <div className="space-y-4">
            <span className="px-3 py-1 bg-slate-200 rounded text-[10px] font-bold uppercase tracking-wider text-slate-800">
              Физическое и цифровое сохранение
            </span>
            <h2 className="text-2xl sm:text-3xl font-black tracking-tight text-slate-950">
              Музей деревянной архитектуры
            </h2>
            <p className="text-sm sm:text-base text-slate-600 leading-relaxed">
              Цифровой архив и коллекция спасенных элементов старинного деревянного зодчества. Фотографии, обмеры и описания резных наличников, дверей и карнизов с дореволюционных усадеб Астрахани.
            </p>
          </div>

          <div className="pt-8 flex flex-wrap items-center gap-3">
            <Link
              href="/projects/wooden-museum"
              className="inline-flex items-center gap-2 px-6 py-3.5 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-bold uppercase tracking-wider transition-colors shadow-xs"
            >
              <Building2 className="w-4 h-4" />
              <span>Открыть каталог музея</span>
            </Link>
          </div>
        </div>

      </div>

    </div>
  );
}
