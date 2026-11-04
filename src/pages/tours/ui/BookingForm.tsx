'use client';

import React, { useState } from 'react';
import { CheckCircle2 } from 'lucide-react';
import type { Tour } from '@/entities/tour';
import { formatRub } from '@/shared/lib';
import { Button, Field, FormSelect, Input, Textarea } from '@/shared/ui';

interface BookingFormProps {
  tours: Tour[]; // только групповые экскурсии
  tourId: string;
  onTourChange: (tourId: string) => void;
}

// Запись на пешеходную экскурсию (ТЗ 10.1). Пока без отправки на сервер.
export function BookingForm({ tours, tourId, onTourChange }: BookingFormProps) {
  const [name, setName] = useState('');
  const [phone, setPhone] = useState('');
  const [email, setEmail] = useState('');
  const [guests, setGuests] = useState('1');
  const [comment, setComment] = useState('');
  const [isSubmitted, setIsSubmitted] = useState(false);

  if (isSubmitted) {
    return (
      <div className="p-8 bg-white border border-slate-200 rounded-2xl text-center space-y-4">
        <CheckCircle2 className="w-12 h-12 text-emerald-600 mx-auto" />
        <h3 className="text-xl font-bold text-slate-900">Вы успешно записаны на экскурсию!</h3>
        <p className="text-sm text-slate-600 max-w-md mx-auto">
          Мы отправили подтверждение на {email || phone}. Ждем вас в назначенное время у точки сбора.
        </p>
        <Button variant="secondary" onClick={() => setIsSubmitted(false)}>
          Записаться на другую дату
        </Button>
      </div>
    );
  }

  return (
    <form
      onSubmit={(e: React.FormEvent) => {
        e.preventDefault();
        setIsSubmitted(true);
      }}
      className="space-y-4"
    >
      <Field label="Выбранная экскурсия" required>
        <FormSelect value={tourId} onChange={(e) => onTourChange(e.target.value)} className="font-medium">
          {tours.map((t) => (
            <option key={t.id} value={t.id}>
              {t.title} ({formatRub(t.price)})
            </option>
          ))}
        </FormSelect>
      </Field>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <Field label="Ваше имя" required>
          <Input required placeholder="Александр" value={name} onChange={(e) => setName(e.target.value)} />
        </Field>
        <Field label="Телефон (для связи в день прогулки)" required>
          <Input type="tel" required placeholder="+7 (999) 000-00-00" value={phone} onChange={(e) => setPhone(e.target.value)} />
        </Field>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <Field label="Электронная почта (Email)" required>
          <Input type="email" required placeholder="Ваш e-mail" value={email} onChange={(e) => setEmail(e.target.value)} />
        </Field>
        <Field label="Количество участников">
          <FormSelect value={guests} onChange={(e) => setGuests(e.target.value)}>
            <option value="1">1 человек</option>
            <option value="2">2 человека</option>
            <option value="3">3 человека</option>
            <option value="4+">4 и более</option>
          </FormSelect>
        </Field>
      </div>

      <Field label="Пожелания или вопросы к экскурсоводу">
        <Textarea
          rows={3}
          placeholder="Например: есть ли скидка для студентов или удобный ли маршрут с детской коляской..."
          value={comment}
          onChange={(e) => setComment(e.target.value)}
        />
      </Field>

      <Button type="submit">Подтвердить запись</Button>
    </form>
  );
}
