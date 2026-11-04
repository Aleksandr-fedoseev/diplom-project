import Link from 'next/link';
import { Building2, Compass } from 'lucide-react';
import { SectionHeading } from './SectionHeading';

// Собственные проекты Центра: карта наследия и музей деревянной архитектуры.
export function OwnProjects() {
  return (
    <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 w-full">
      <SectionHeading kicker="Инициативы фонда" title="Собственные проекты" />
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
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
            <h3 className="text-xl sm:text-2xl font-black tracking-tight text-white">Интерактивная карта наследия</h3>
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

        <div className="bg-slate-50 border border-slate-200 rounded-3xl p-8 flex flex-col justify-between group">
          <div className="space-y-2">
            <span className="px-2.5 py-0.5 bg-slate-200 rounded text-[10px] font-bold uppercase tracking-wider text-slate-800">
              Проект №2
            </span>
            <h3 className="text-xl sm:text-2xl font-black tracking-tight text-slate-950">Музей деревянной архитектуры</h3>
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
            <Link href="/support?tab=wooden" className="text-xs font-bold text-slate-600 hover:text-slate-950 transition-colors">
              Передать наличник →
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}
