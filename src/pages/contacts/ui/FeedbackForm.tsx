'use client';

import React, { useState } from 'react';
import { CheckCircle2, MessageSquare, Send } from 'lucide-react';
import { Field, FormSelect, Input, Textarea } from '@/shared/ui';

const TOPICS = [
  { value: 'general', label: 'Общий вопрос или предложение' },
  { value: 'story', label: 'Предложить тему или материал для расследования' },
  { value: 'archive', label: 'Передать архивные фотографии или документы' },
  { value: 'wooden', label: 'Передать резной наличник или элемент в музей' },
  { value: 'tours', label: 'Вопрос по пешеходным экскурсиям' },
  { value: 'partnership', label: 'Партнерство и совместные проекты' },
];

const EMPTY = { name: '', contact: '', topic: 'general', message: '' };

// Форма обратной связи (ТЗ 11). Пока без отправки на сервер.
export function FeedbackForm() {
  const [form, setForm] = useState(EMPTY);
  const [submitted, setSubmitted] = useState(false);
  const set = (patch: Partial<typeof EMPTY>) => setForm((f) => ({ ...f, ...patch }));

  return (
    <div className="bg-slate-50 border border-slate-200 rounded-3xl p-6 sm:p-8 lg:p-10">
      <div className="space-y-2 mb-6">
        <div className="inline-flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider text-slate-500">
          <MessageSquare className="w-4 h-4" />
          <span>Обратная связь</span>
        </div>
        <h2 className="text-xl sm:text-2xl font-black text-slate-950 tracking-tight">Форма обратной связи</h2>
        <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
          Отправьте вопрос, предложение по теме статьи или информацию о сохранившихся архивных материалах:
        </p>
      </div>

      {submitted ? (
        <div className="p-8 bg-white border border-emerald-200 rounded-2xl text-center space-y-3">
          <div className="w-12 h-12 rounded-full bg-emerald-50 text-emerald-600 flex items-center justify-center mx-auto">
            <CheckCircle2 className="w-6 h-6" />
          </div>
          <h3 className="text-base font-bold text-slate-950">Сообщение успешно отправлено!</h3>
          <p className="text-xs text-slate-600 max-w-sm mx-auto">
            Спасибо за обращение. Редакция и координаторы Центра сохранения архитектуры свяжутся с вами в ближайшее время.
          </p>
          <button
            type="button"
            onClick={() => {
              setSubmitted(false);
              setForm(EMPTY);
            }}
            className="mt-4 px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-900 rounded-xl text-xs font-bold uppercase tracking-wider transition-colors"
          >
            Отправить ещё сообщение
          </button>
        </div>
      ) : (
        <form
          onSubmit={(e: React.FormEvent) => {
            e.preventDefault();
            setSubmitted(true);
          }}
          className="space-y-4"
        >
          <Field label="Ваше имя" required>
            <Input required placeholder="Как к вам обращаться" value={form.name} onChange={(e) => set({ name: e.target.value })} />
          </Field>
          <Field label="Контакт для связи (Telegram, телефон или e-mail)" required>
            <Input required placeholder="@username или телефон для связи" value={form.contact} onChange={(e) => set({ contact: e.target.value })} />
          </Field>
          <Field label="Тема обращения">
            <FormSelect value={form.topic} onChange={(e) => set({ topic: e.target.value })}>
              {TOPICS.map((t) => (
                <option key={t.value} value={t.value}>
                  {t.label}
                </option>
              ))}
            </FormSelect>
          </Field>
          <Field label="Текст сообщения" required>
            <Textarea
              required
              className="resize-none"
              placeholder="Опишите ваш вопрос, предложение или информацию об объекте..."
              value={form.message}
              onChange={(e) => set({ message: e.target.value })}
            />
          </Field>
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
  );
}
