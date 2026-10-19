import React from 'react';
import Link from 'next/link';
import { 
  ArrowRight, 
  Compass, 
  Building2, 
  Upload, 
  MessageSquare, 
  Coins, 
  Heart
} from 'lucide-react';
import { STORIES } from '@/data/stories';

export default function HomePage() {
  const leadStory = STORIES.find(s => s.isLeadArticle) || STORIES[0];
  const recentStories = STORIES.filter(s => s.id !== leadStory.id).slice(0, 3);

  return (
    <div className="flex flex-col gap-12 sm:gap-16 pb-20 bg-white text-slate-900">
      
      {/* ========================================================================= */}
      {/* Главный фокусный материал (Hero)                                          */}
      {/* ========================================================================= */}
      <section className="border-b border-slate-200 bg-slate-50/50 pt-6 pb-10 sm:pt-8 sm:pb-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="bg-slate-950 rounded-3xl text-white overflow-hidden relative shadow-lg min-h-[400px] sm:min-h-[460px] p-6 sm:p-10 lg:p-12 flex flex-col justify-between">
            <div 
              className="absolute inset-0 bg-cover bg-center opacity-40 mix-blend-luminosity"
              style={{ backgroundImage: `url(${leadStory.coverImage})` }}
            />
            <div className="absolute inset-0 bg-gradient-to-t from-slate-950 via-slate-950/60 to-transparent" />

            <div className="relative z-10 space-y-4 max-w-3xl">
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded bg-white/20 backdrop-blur-md text-xs font-bold uppercase tracking-wider text-white">
                <span>{leadStory.categoryLabel}</span>
                <span>•</span>
                <span>{leadStory.publishedAt}</span>
                <span>•</span>
                <span>{leadStory.readTime}</span>
              </div>

              <h1 className="text-2xl sm:text-4xl lg:text-5xl font-black tracking-tight leading-tight text-white">
                {leadStory.title}
              </h1>

              <p className="text-sm sm:text-base lg:text-lg text-slate-200 leading-relaxed font-normal">
                {leadStory.lead}
              </p>
            </div>

            <div className="relative z-10 pt-6 flex flex-wrap items-center gap-4">
              <Link
                href={`/stories/${leadStory.slug}`}
                className="inline-flex items-center gap-2 px-6 py-3.5 bg-white hover:bg-slate-100 text-slate-950 text-xs sm:text-sm font-bold uppercase tracking-wider rounded-xl transition-all shadow-md"
              >
                <span>Читать статью</span>
                <ArrowRight className="w-4 h-4" />
              </Link>

              <Link
                href="/map"
                className="inline-flex items-center gap-2 px-5 py-3.5 bg-slate-800/80 hover:bg-slate-800 text-white text-xs sm:text-sm font-semibold rounded-xl backdrop-blur-md transition-all border border-slate-700"
              >
                <Compass className="w-4 h-4" />
                <span>Открыть на карте</span>
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* ========================================================================= */}
      {/* Свежие публикации (Картотека расследований)                                */}
      {/* ========================================================================= */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 w-full">
        <div className="flex items-end justify-between mb-8">
          <div>
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
              Краеведческое медиа
            </span>
            <h2 className="text-2xl sm:text-3xl font-black text-slate-950 tracking-tight mt-1">
              Свежие материалы
            </h2>
          </div>
          <Link
            href="/stories"
            className="text-xs sm:text-sm font-bold text-slate-900 hover:text-blue-600 flex items-center gap-1 transition-colors"
          >
            <span>Все истории</span>
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {recentStories.map((story) => (
            <Link
              key={story.id}
              href={`/stories/${story.slug}`}
              className="group bg-white border border-slate-200 rounded-2xl overflow-hidden shadow-xs hover:shadow-md hover:border-slate-300 transition-all flex flex-col justify-between"
            >
              <div>
                <div className="h-48 bg-slate-100 overflow-hidden relative">
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

                  <p className="text-xs text-slate-600 line-clamp-2 leading-relaxed">
                    {story.lead}
                  </p>
                </div>
              </div>

              <div className="px-5 pb-5 pt-0 flex items-center justify-between text-xs font-semibold text-slate-900 group-hover:text-blue-600">
                <span className="text-[11px] text-slate-400 font-normal">{story.author.name}</span>
                <span className="flex items-center gap-1">
                  <span>Читать</span>
                  <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition-transform" />
                </span>
              </div>
            </Link>
          ))}
        </div>
      </section>

      {/* ========================================================================= */}
      {/* Ключевые проекты (Карта и Музей деревянного зодчества)                     */}
      {/* ========================================================================= */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 w-full">
        <div className="mb-6">
          <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
            Инициативы фонда
          </span>
          <h2 className="text-2xl sm:text-3xl font-black text-slate-950 tracking-tight mt-1">
            Собственные проекты
          </h2>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          
          {/* Проект 1: Интерактивная карта */}
          <div className="bg-slate-950 text-white rounded-3xl p-8 flex flex-col justify-between overflow-hidden relative group shadow-md">
            <div 
              className="absolute inset-0 bg-cover bg-center opacity-25 group-hover:scale-105 transition-transform duration-500"
              style={{ backgroundImage: 'url(/heritage/birzha/cover.jpg)' }}
            />
            <div className="absolute inset-0 bg-gradient-to-t from-slate-950 via-slate-950/70 to-transparent" />

            <div className="relative z-10 space-y-2">
              <span className="px-2.5 py-0.5 bg-white/20 backdrop-blur-md rounded text-[10px] font-bold uppercase tracking-wider text-white">
                Проект №1
              </span>
              <h3 className="text-xl sm:text-2xl font-black tracking-tight text-white">
                Интерактивная карта наследия
              </h3>
              <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
                Каталог архитектурных памятников Астрахани с архивными фото и описаниями на карте города.
              </p>
            </div>

            <div className="relative z-10 pt-6 flex items-center">
              <Link
                href="/map"
                className="inline-flex items-center gap-2 px-5 py-3 bg-white hover:bg-slate-100 text-slate-950 rounded-xl text-xs font-bold uppercase tracking-wider transition-colors shadow-sm"
              >
                <Compass className="w-4 h-4" />
                <span>Открыть карту</span>
              </Link>
            </div>
          </div>

          {/* Проект 2: Музей деревянной архитектуры */}
          <div className="bg-slate-50 border border-slate-200 rounded-3xl p-8 flex flex-col justify-between group">
            <div className="space-y-2">
              <span className="px-2.5 py-0.5 bg-slate-200 rounded text-[10px] font-bold uppercase tracking-wider text-slate-800">
                Проект №2
              </span>
              <h3 className="text-xl sm:text-2xl font-black tracking-tight text-slate-950">
                Музей деревянной архитектуры
              </h3>
              <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
                Сохранение и цифровой архив резных наличников, карнизов и элементов дореволюционных усадеб.
              </p>
            </div>

            <div className="pt-6 flex flex-wrap items-center gap-3">
              <Link
                href="/projects/wooden-museum"
                className="inline-flex items-center gap-2 px-5 py-3 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-bold uppercase tracking-wider transition-colors shadow-xs"
              >
                <Building2 className="w-4 h-4" />
                <span>Каталог элементов</span>
              </Link>
              <Link
                href="/support#wooden"
                className="text-xs font-bold text-slate-600 hover:text-slate-950 transition-colors"
              >
                Передать наличник →
              </Link>
            </div>
          </div>

        </div>
      </section>

      {/* ========================================================================= */}
      {/* Помощь проекту и участие жителей                                          */}
      {/* ========================================================================= */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 w-full">
        <div className="bg-slate-950 text-white rounded-3xl p-8 sm:p-10 border border-slate-900">
          <div className="max-w-2xl space-y-2 mb-8">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded bg-white/20 backdrop-blur-md text-[11px] font-bold uppercase tracking-wider text-white">
              <Heart className="w-3.5 h-3.5 text-rose-400 fill-rose-400" />
              <span>Участие жителей</span>
            </div>
            <h2 className="text-2xl sm:text-3xl font-black text-white tracking-tight">
              Помогите сохранить историю города
            </h2>
            <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
              Вы можете внести свой вклад в сохранение культурного наследия Астрахани:
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            
            {/* 1. Пожертвование */}
            <Link
              href="/support?tab=money"
              className="p-5 bg-slate-900 hover:bg-slate-850 border border-slate-800 rounded-2xl transition-all group flex flex-col justify-between"
            >
              <div className="space-y-2">
                <Coins className="w-5 h-5 text-amber-400" />
                <h4 className="text-sm font-bold text-white group-hover:text-amber-300 transition-colors">
                  Пожертвование
                </h4>
                <p className="text-xs text-slate-400 leading-relaxed">
                  На закупку составов для консервации дерева и архивные копии.
                </p>
              </div>
              <span className="pt-4 text-xs font-bold text-white flex items-center gap-1">
                <span>Поддержать</span>
                <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition-transform" />
              </span>
            </Link>

            {/* 2. Передать элемент резьбы */}
            <Link
              href="/support?tab=wooden"
              className="p-5 bg-slate-900 hover:bg-slate-850 border border-slate-800 rounded-2xl transition-all group flex flex-col justify-between"
            >
              <div className="space-y-2">
                <Building2 className="w-5 h-5 text-emerald-400" />
                <h4 className="text-sm font-bold text-white group-hover:text-emerald-300 transition-colors">
                  Передать деталь
                </h4>
                <p className="text-xs text-slate-400 leading-relaxed">
                  Старый наличник, дверь или карниз для музея.
                </p>
              </div>
              <span className="pt-4 text-xs font-bold text-white flex items-center gap-1">
                <span>В музей</span>
                <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition-transform" />
              </span>
            </Link>

            {/* 3. Архивные материалы */}
            <Link
              href="/support?tab=archive"
              className="p-5 bg-slate-900 hover:bg-slate-850 border border-slate-800 rounded-2xl transition-all group flex flex-col justify-between"
            >
              <div className="space-y-2">
                <Upload className="w-5 h-5 text-blue-400" />
                <h4 className="text-sm font-bold text-white group-hover:text-blue-300 transition-colors">
                  Архивные фото
                </h4>
                <p className="text-xs text-slate-400 leading-relaxed">
                  Снимки, открытки и документы из семейных архивов.
                </p>
              </div>
              <span className="pt-4 text-xs font-bold text-white flex items-center gap-1">
                <span>Прислать фото</span>
                <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition-transform" />
              </span>
            </Link>

            {/* 4. Рассказать историю */}
            <Link
              href="/support?tab=story"
              className="p-5 bg-slate-900 hover:bg-slate-850 border border-slate-800 rounded-2xl transition-all group flex flex-col justify-between"
            >
              <div className="space-y-2">
                <MessageSquare className="w-5 h-5 text-purple-400" />
                <h4 className="text-sm font-bold text-white group-hover:text-purple-300 transition-colors">
                  История дома
                </h4>
                <p className="text-xs text-slate-400 leading-relaxed">
                  Воспоминания о семье или старинном здании.
                </p>
              </div>
              <span className="pt-4 text-xs font-bold text-white flex items-center gap-1">
                <span>Поделиться</span>
                <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition-transform" />
              </span>
            </Link>

          </div>
        </div>
      </section>

    </div>
  );
}
