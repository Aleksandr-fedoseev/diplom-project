'use client';

import { useState } from 'react';
import { TOURS } from '@/entities/tour';
import { BookingForm } from './BookingForm';
import { TourCard } from './TourCard';

const GROUP_TOURS = TOURS.filter((t) => !t.isOnlineRoute);

export default function ToursPage() {
  const [tourId, setTourId] = useState(GROUP_TOURS[0]?.id ?? '');

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      <div className="mb-10">
        <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Городские экспедиции</span>
        <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-slate-950 mt-1">Экскурсии и маршруты</h1>
        <p className="text-base sm:text-lg text-slate-600 mt-2 max-w-2xl">
          Пешеходные прогулки с историками Центра и бесплатные онлайн-маршруты по карте для самостоятельного изучения
          Астрахани.
        </p>
      </div>

      <div className="space-y-8 mb-16">
        {TOURS.map((tour) => (
          <TourCard key={tour.id} tour={tour} onBook={setTourId} />
        ))}
      </div>

      {GROUP_TOURS.length > 0 && (
        <section id="booking" className="scroll-mt-24 max-w-3xl mx-auto bg-slate-50 border border-slate-200 rounded-3xl p-6 sm:p-10 shadow-xs">
          <div className="mb-8 text-center space-y-2">
            <h2 className="text-2xl sm:text-3xl font-black text-slate-950">Запись на пешеходную экскурсию</h2>
            <p className="text-xs sm:text-sm text-slate-600">
              Заполните контакты, и координатор группы свяжется с вами для подтверждения брони.
            </p>
          </div>
          <BookingForm tours={GROUP_TOURS} tourId={tourId} onTourChange={setTourId} />
        </section>
      )}
    </div>
  );
}
