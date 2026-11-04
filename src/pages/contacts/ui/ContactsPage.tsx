import Link from 'next/link';
import { Building2, Clock, Compass, ShoppingBag, Users } from 'lucide-react';
import { SOCIAL_LINKS } from '@/shared/config';
import { FeedbackForm } from './FeedbackForm';

export default function ContactsPage() {
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      <div className="mb-10">
        <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Связь с командой проекта</span>
        <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-slate-950 mt-1">Где нас найти</h1>
        <p className="text-sm sm:text-base text-slate-600 mt-2 max-w-2xl">
          Информация о взаимодействии с АНО «Центр сохранения архитектуры», партнерских площадках и форма обратной связи.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-12 items-start">
        <div className="lg:col-span-6 space-y-6">
          <section className="bg-slate-50 border border-slate-200 rounded-3xl p-6 sm:p-8 space-y-4">
            <h2 className="flex items-center gap-2 text-base font-bold text-slate-950">
              <Building2 className="w-5 h-5 text-slate-900" />
              <span>Штаб и исследовательская мастерская</span>
            </h2>
            <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
              Проект развивается силами историков, реставраторов и волонтеров АНО «Центр сохранения архитектуры». Прием
              архивных материалов, очные встречи исследователей и передача элементов деревянного зодчества осуществляются
              по предварительной договоренности.
            </p>
            <dl className="space-y-3 pt-2 text-xs sm:text-sm text-slate-700">
              <div className="flex items-start gap-3">
                <Clock className="w-4 h-4 text-slate-500 shrink-0 mt-0.5" />
                <div>
                  <dt className="font-bold text-slate-900">Режим работы:</dt>
                  <dd className="text-slate-500">Среда – Воскресенье (по предварительному согласованию)</dd>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <Users className="w-4 h-4 text-slate-500 shrink-0 mt-0.5" />
                <div>
                  <dt className="font-bold text-slate-900">Городские лектории и встречи:</dt>
                  <dd className="text-slate-500">Анонсируются в календаре событий и официальных каналах сообщества</dd>
                </div>
              </div>
            </dl>
            <div className="pt-4 border-t border-slate-200 flex flex-wrap gap-3">
              {SOCIAL_LINKS.map((link) => (
                <a
                  key={link.href}
                  href={link.href}
                  target="_blank"
                  rel="noreferrer"
                  className="flex-1 min-w-[140px] py-2.5 px-4 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-bold uppercase tracking-wider text-center transition-colors"
                >
                  {link.label}
                </a>
              ))}
            </div>
          </section>

          <section className="bg-white border border-slate-200 rounded-3xl p-6 sm:p-8 space-y-4">
            <h2 className="flex items-center gap-2 text-base font-bold text-slate-950">
              <ShoppingBag className="w-5 h-5 text-slate-900" />
              <span>Точки распространения мерча и изданий</span>
            </h2>
            <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
              Печатные каталоги памятников, книги и краеведческие издания фонда доступны к заказу онлайн на сайте, а также
              на стартах авторских экскурсий и просветительских мероприятий.
            </p>
            <Link href="/shop" className="inline-flex text-xs font-bold text-slate-900 hover:text-blue-600 transition-colors uppercase tracking-wider">
              Перейти в каталог мерча →
            </Link>
          </section>

          <section className="bg-slate-950 text-white rounded-3xl p-6 sm:p-8 flex flex-col justify-between space-y-4">
            <div className="space-y-2">
              <span className="px-2.5 py-0.5 bg-white/20 rounded text-[10px] font-bold uppercase tracking-wider text-white">
                Геолокация объектов
              </span>
              <h2 className="text-lg sm:text-xl font-bold text-white">Интерактивная карта памятников города</h2>
              <p className="text-xs text-slate-300 leading-relaxed">
                Все архитектурные памятники, купеческие особняки и маршруты пешеходных экскурсий нанесены на карту с
                точными координатами.
              </p>
            </div>
            <div>
              <Link
                href="/map"
                className="inline-flex items-center gap-2 px-5 py-2.5 bg-white hover:bg-slate-100 text-slate-950 rounded-xl text-xs font-bold uppercase tracking-wider transition-colors"
              >
                <Compass className="w-4 h-4" />
                <span>Открыть карту города</span>
              </Link>
            </div>
          </section>
        </div>

        <div className="lg:col-span-6">
          <FeedbackForm />
        </div>
      </div>
    </div>
  );
}
