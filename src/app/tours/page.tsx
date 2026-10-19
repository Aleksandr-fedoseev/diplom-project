'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { 
  Compass, 
  Clock, 
  MapPin, 
  Users, 
  Calendar, 
  Volume2, 
  ArrowRight, 
  CheckCircle2, 
  Phone, 
  Mail,
  UserCheck
} from 'lucide-react';
import { TOURS } from '@/data/tours';

export default function ToursPage() {
  const [selectedTour, setSelectedTour] = useState(TOURS[0].title);
  const [name, setName] = useState('');
  const [phone, setPhone] = useState('');
  const [email, setEmail] = useState('');
  const [guestsCount, setGuestsCount] = useState('1');
  const [comment, setComment] = useState('');
  const [isSubmitted, setIsSubmitted] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitted(true);
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      
      {/* Header */}
      <div className="mb-10">
        <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
          Городские экспедиции
        </span>
        <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-slate-950 mt-1">
          Экскурсии и маршруты
        </h1>
        <p className="text-base sm:text-lg text-slate-600 mt-2 max-w-2xl">
          Пешеходные прогулки с историками Центра и бесплатные онлайн-маршруты с аудиогидом для самостоятельного изучения Астрахани.
        </p>
      </div>

      {/* Tours Grid */}
      <div className="space-y-8 mb-16">
        {TOURS.map((tour) => (
          <div
            key={tour.id}
            id={tour.slug}
            className="bg-white border border-slate-200 rounded-3xl overflow-hidden shadow-xs hover:shadow-md transition-all grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch"
          >
            {/* Image Col (1-5) */}
            <div className="lg:col-span-5 h-64 lg:h-auto bg-slate-100 relative">
              <img
                src={tour.coverImage}
                alt={tour.title}
                className="w-full h-full object-cover"
              />
              <div className="absolute top-4 left-4 flex gap-2">
                <span className="px-3 py-1 bg-slate-950/85 backdrop-blur-md text-white text-xs font-bold uppercase tracking-wider rounded-lg">
                  {tour.isOnlineRoute ? 'Онлайн-аудиогид' : 'Пешеходная группа'}
                </span>
              </div>
            </div>

            {/* Info Col (6-12) */}
            <div className="lg:col-span-7 p-6 sm:p-8 flex flex-col justify-between space-y-6">
              <div className="space-y-4">
                <div className="flex flex-wrap items-center gap-3 text-xs font-semibold text-slate-500">
                  <span className="flex items-center gap-1.5 bg-slate-100 px-2.5 py-1 rounded text-slate-800">
                    <Clock className="w-3.5 h-3.5" />
                    <span>{tour.duration}</span>
                  </span>
                  <span>•</span>
                  <span>{tour.distance}</span>
                  <span>•</span>
                  <span>{tour.stopsCount} остановок</span>
                  {tour.nextDate && (
                    <>
                      <span>•</span>
                      <span className="text-slate-900 font-bold bg-amber-50 border border-amber-200 px-2 py-0.5 rounded">
                        Ближайшая: {tour.nextDate}
                      </span>
                    </>
                  )}
                </div>

                <h3 className="text-2xl font-bold text-slate-950 leading-snug">
                  {tour.title}
                </h3>

                <p className="text-sm text-slate-600 leading-relaxed">
                  {tour.description}
                </p>

                {/* Stops sequence */}
                <div className="pt-2">
                  <h4 className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-2">
                    Ключевые точки маршрута:
                  </h4>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
                    {tour.stops.map((stop, i) => (
                      <div key={i} className="flex items-start gap-2 p-2 bg-slate-50 rounded-lg border border-slate-100">
                        <span className="w-4 h-4 rounded-full bg-slate-900 text-white flex items-center justify-center font-bold text-[9px] shrink-0 mt-0.5">
                          {i + 1}
                        </span>
                        <div>
                          <strong className="block text-slate-900 leading-tight">{stop.title}</strong>
                          <span className="text-slate-500 text-[10px]">{stop.address}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {tour.guideName && (
                  <div className="pt-2 text-xs text-slate-500 flex items-center gap-2">
                    <UserCheck className="w-4 h-4 text-slate-700" />
                    <span>Ведущий: <strong>{tour.guideName}</strong> ({tour.guideRole})</span>
                  </div>
                )}
              </div>

              {/* Price & Action */}
              <div className="pt-4 border-t border-slate-100 flex flex-wrap items-center justify-between gap-4">
                <div>
                  <span className="text-[11px] text-slate-400 uppercase tracking-wider block font-bold">Стоимость:</span>
                  <span className="text-2xl font-black text-slate-950">
                    {tour.price === 0 ? 'Бесплатно' : `${tour.price} ₽`}
                  </span>
                </div>

                {tour.isOnlineRoute ? (
                  <Link
                    href="/map"
                    className="inline-flex items-center gap-2 px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-xl text-xs font-bold uppercase tracking-wider transition-colors shadow-sm"
                  >
                    <Volume2 className="w-4 h-4" />
                    <span>Слушать онлайн на карте</span>
                  </Link>
                ) : (
                  <a
                    href="#booking"
                    onClick={() => setSelectedTour(tour.title)}
                    className="inline-flex items-center gap-2 px-6 py-3 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-bold uppercase tracking-wider transition-colors shadow-xs"
                  >
                    <span>Записаться в группу</span>
                    <ArrowRight className="w-4 h-4" />
                  </a>
                )}
              </div>

            </div>
          </div>
        ))}
      </div>

      {/* Booking Form (ТЗ 10.1: Форма записи на экскурсию) */}
      <section id="booking" className="scroll-mt-24 max-w-3xl mx-auto bg-slate-50 border border-slate-200 rounded-3xl p-6 sm:p-10 shadow-xs">
        <div className="mb-8 text-center space-y-2">
          <h2 className="text-2xl sm:text-3xl font-black text-slate-950">
            Запись на пешеходную экскурсию
          </h2>
          <p className="text-xs sm:text-sm text-slate-600">
            Заполните контакты, и координатор группы свяжется с вами для подтверждения брони.
          </p>
        </div>

        {isSubmitted ? (
          <div className="p-8 bg-white border border-slate-200 rounded-2xl text-center space-y-4">
            <CheckCircle2 className="w-12 h-12 text-emerald-600 mx-auto" />
            <h3 className="text-xl font-bold text-slate-900">
              Вы успешно записаны на экскурсию!
            </h3>
            <p className="text-sm text-slate-600 max-w-md mx-auto">
              Мы отправили подтверждение на {email || phone}. Ждем вас в назначенное время у точки сбора.
            </p>
            <button
              onClick={() => setIsSubmitted(false)}
              className="px-6 py-2.5 bg-slate-950 text-white rounded-xl text-xs font-bold uppercase tracking-wider"
            >
              Записаться на другую дату
            </button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                Выбранная экскурсия*
              </label>
              <select
                value={selectedTour}
                onChange={(e) => setSelectedTour(e.target.value)}
                className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950 font-medium"
              >
                {TOURS.filter(t => !t.isOnlineRoute).map(t => (
                  <option key={t.id} value={t.title}>{t.title} ({t.price} ₽)</option>
                ))}
              </select>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                  Ваше имя*
                </label>
                <input
                  type="text"
                  required
                  placeholder="Александр"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                />
              </div>

              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                  Телефон (для связи в день прогулки)*
                </label>
                <input
                  type="tel"
                  required
                  placeholder="+7 (999) 000-00-00"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                />
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                  Электронная почта (Email)*
                </label>
                <input
                  type="email"
                  required
                  placeholder="Ваш e-mail"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                />
              </div>

              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                  Количество участников
                </label>
                <select
                  value={guestsCount}
                  onChange={(e) => setGuestsCount(e.target.value)}
                  className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                >
                  <option value="1">1 человек</option>
                  <option value="2">2 человека</option>
                  <option value="3">3 человека</option>
                  <option value="4+">4 и более</option>
                </select>
              </div>
            </div>

            <div>
              <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                Пожелания или вопросы к экскурсоводу
              </label>
              <textarea
                rows={3}
                placeholder="Например: есть ли скидка для студентов или удобный ли маршрут с детской коляской..."
                value={comment}
                onChange={(e) => setComment(e.target.value)}
                className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
              />
            </div>

            <button
              type="submit"
              className="w-full py-4 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-black uppercase tracking-wider shadow-sm transition-colors"
            >
              Подтвердить запись
            </button>
          </form>
        )}
      </section>

    </div>
  );
}
