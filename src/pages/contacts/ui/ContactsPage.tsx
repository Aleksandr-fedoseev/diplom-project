'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { 
  MapPin, 
  Clock, 
  Compass, 
  ShoppingBag, 
  Send, 
  MessageSquare,
  CheckCircle2,
  Building2,
  Users
} from 'lucide-react';

export default function ContactsPage() {
  const [submitted, setSubmitted] = useState(false);
  const [form, setForm] = useState({
    name: '',
    contact: '',
    topic: 'general',
    message: '',
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitted(true);
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      
      {/* 11. Заголовок раздела «Где нас найти» */}
      <div className="mb-10">
        <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
          Связь с командой проекта
        </span>
        <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-slate-950 mt-1">
          Где нас найти
        </h1>
        <p className="text-sm sm:text-base text-slate-600 mt-2 max-w-2xl">
          Информация о взаимодействии с АНО «Центр сохранения архитектуры», партнерских площадках и форма обратной связи.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-12 items-start">
        
        {/* Левая колонка: Информация о точках проекта и официальных каналах (Col 1-6) */}
        <div className="lg:col-span-6 space-y-6">
          
          {/* Штаб и формат работы */}
          <div className="bg-slate-50 border border-slate-200 rounded-3xl p-6 sm:p-8 space-y-4">
            <div className="flex items-center gap-2">
              <Building2 className="w-5 h-5 text-slate-900" />
              <h3 className="text-base font-bold text-slate-950">
                Штаб и исследовательская мастерская
              </h3>
            </div>

            <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
              Проект развивается силами историков, реставраторов и волонтеров АНО «Центр сохранения архитектуры». Прием архивных материалов, очные встречи исследователей и передача элементов деревянного зодчества осуществляются по предварительной договоренности.
            </p>

            <div className="space-y-3 pt-2 text-xs sm:text-sm text-slate-700">
              <div className="flex items-start gap-3">
                <Clock className="w-4 h-4 text-slate-500 shrink-0 mt-0.5" />
                <div>
                  <strong className="block text-slate-900">Режим работы:</strong>
                  <span className="text-slate-500">Среда – Воскресенье (по предварительному согласованию)</span>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <Users className="w-4 h-4 text-slate-500 shrink-0 mt-0.5" />
                <div>
                  <strong className="block text-slate-900">Городские лектории и встречи:</strong>
                  <span className="text-slate-500">Анонсируются в календаре событий и официальных каналах сообщества</span>
                </div>
              </div>
            </div>

            <div className="pt-4 border-t border-slate-200 flex flex-wrap gap-3">
              <a
                href="https://vk.com"
                target="_blank"
                rel="noreferrer"
                className="flex-1 min-w-[140px] py-2.5 px-4 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-bold uppercase tracking-wider text-center transition-colors"
              >
                ВКонтакте
              </a>
              <a
                href="https://t.me"
                target="_blank"
                rel="noreferrer"
                className="flex-1 min-w-[140px] py-2.5 px-4 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-bold uppercase tracking-wider text-center transition-colors"
              >
                Telegram-канал
              </a>
            </div>
          </div>

          {/* Партнерские площадки и распространение мерча */}
          <div className="bg-white border border-slate-200 rounded-3xl p-6 sm:p-8 space-y-4">
            <div className="flex items-center gap-2">
              <ShoppingBag className="w-5 h-5 text-slate-900" />
              <h3 className="text-base font-bold text-slate-950">
                Точки распространения мерча и изданий
              </h3>
            </div>

            <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
              Печатные каталоги памятников, книги и краеведческие издания фонда доступны к заказу онлайн на сайте, а также на стартах авторских экскурсий и просветительских мероприятий.
            </p>

            <div className="pt-2">
              <Link
                href="/shop"
                className="inline-flex items-center gap-2 text-xs font-bold text-slate-900 hover:text-blue-600 transition-colors uppercase tracking-wider"
              >
                <span>Перейти в каталог мерча →</span>
              </Link>
            </div>
          </div>

          {/* Ссылка на карту */}
          <div className="bg-slate-950 text-white rounded-3xl p-6 sm:p-8 flex flex-col justify-between space-y-4">
            <div className="space-y-2">
              <span className="px-2.5 py-0.5 bg-white/20 rounded text-[10px] font-bold uppercase tracking-wider text-white">
                Геолокация объектов
              </span>
              <h4 className="text-lg sm:text-xl font-bold text-white">
                Интерактивная карта памятников города
              </h4>
              <p className="text-xs text-slate-300 leading-relaxed">
                Все архитектурные памятники, купеческие особняки и маршруты пешеходных экскурсий нанесены на карту с точными координатами.
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
          </div>

        </div>

        {/* Правая колонка: Форма обратной связи для вопросов и предложений (ТЗ 11) (Col 7-12) */}
        <div className="lg:col-span-6 bg-slate-50 border border-slate-200 rounded-3xl p-6 sm:p-8 lg:p-10">
          <div className="space-y-2 mb-6">
            <div className="inline-flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider text-slate-500">
              <MessageSquare className="w-4 h-4" />
              <span>Пункт 11 ТЗ</span>
            </div>
            <h3 className="text-xl sm:text-2xl font-black text-slate-950 tracking-tight">
              Форма обратной связи
            </h3>
            <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
              Отправьте вопрос, предложение по теме статьи или информацию о сохранившихся архивных материалах:
            </p>
          </div>

          {submitted ? (
            <div className="p-8 bg-white border border-emerald-200 rounded-2xl text-center space-y-3">
              <div className="w-12 h-12 rounded-full bg-emerald-50 text-emerald-600 flex items-center justify-center mx-auto">
                <CheckCircle2 className="w-6 h-6" />
              </div>
              <h4 className="text-base font-bold text-slate-950">
                Сообщение успешно отправлено!
              </h4>
              <p className="text-xs text-slate-600 max-w-sm mx-auto">
                Спасибо за обращение. Редакция и координаторы Центра сохранения архитектуры свяжутся с вами в ближайшее время.
              </p>
              <button
                onClick={() => {
                  setSubmitted(false);
                  setForm({ name: '', contact: '', topic: 'general', message: '' });
                }}
                className="mt-4 px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-900 rounded-xl text-xs font-bold uppercase tracking-wider transition-colors"
              >
                Отправить ещё сообщение
              </button>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1.5">
                  Ваше имя *
                </label>
                <input
                  type="text"
                  required
                  placeholder="Как к вам обращаться"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  className="w-full bg-white border border-slate-200 px-4 py-2.5 rounded-xl text-xs sm:text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                />
              </div>

              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1.5">
                  Контакт для связи (Telegram, телефон или e-mail) *
                </label>
                <input
                  type="text"
                  required
                  placeholder="@username или телефон для связи"
                  value={form.contact}
                  onChange={(e) => setForm({ ...form, contact: e.target.value })}
                  className="w-full bg-white border border-slate-200 px-4 py-2.5 rounded-xl text-xs sm:text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                />
              </div>

              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1.5">
                  Тема обращения
                </label>
                <select
                  value={form.topic}
                  onChange={(e) => setForm({ ...form, topic: e.target.value })}
                  className="w-full bg-white border border-slate-200 px-4 py-2.5 rounded-xl text-xs sm:text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                >
                  <option value="general">Общий вопрос или предложение</option>
                  <option value="story">Предложить тему или материал для расследования</option>
                  <option value="archive">Передать архивные фотографии или документы</option>
                  <option value="wooden">Передать резной наличник или элемент в музей</option>
                  <option value="tours">Вопрос по пешеходным экскурсиям</option>
                  <option value="partnership">Партнерство и совместные проекты</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1.5">
                  Текст сообщения *
                </label>
                <textarea
                  required
                  rows={4}
                  placeholder="Опишите ваш вопрос, предложение или информацию об объекте..."
                  value={form.message}
                  onChange={(e) => setForm({ ...form, message: e.target.value })}
                  className="w-full bg-white border border-slate-200 px-4 py-2.5 rounded-xl text-xs sm:text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950 resize-none"
                />
              </div>

              <button
                type="submit"
                className="w-full py-3.5 px-6 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs sm:text-sm font-bold uppercase tracking-wider transition-colors flex items-center justify-center gap-2 shadow-xs"
              >
                <Send className="w-4 h-4" />
                <span>Отправить сообщение</span>
              </button>

              <p className="text-[11px] text-slate-500 text-center leading-relaxed">
                Нажимая кнопку, вы соглашаетесь на обработку обращения координаторами проекта.
              </p>
            </form>
          )}

        </div>

      </div>

    </div>
  );
}
