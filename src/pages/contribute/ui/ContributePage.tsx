'use client';

import React, { useId, useState } from 'react';
import Link from 'next/link';
import { ArrowLeft, ArrowRight, CheckCircle2, Info, LogIn, Upload, UserCheck } from 'lucide-react';
import { PageHeader } from '@/shared/ui';
import { CATEGORY_OPTIONS, EMPTY_DRAFT, STEPS, type ObjectDraft } from '../model/draft';
import { backClass, DraftInput, DraftSelect, DraftTextarea, Label, nextClass, submitClass } from './fields';
import { Stepper } from './Stepper';

export default function ContributePage() {
  const [step, setStep] = useState(1);
  const [draft, setDraft] = useState<ObjectDraft>(EMPTY_DRAFT);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const id = useId();
  const set = (patch: Partial<ObjectDraft>) => setDraft((d) => ({ ...d, ...patch }));

  // Каждый шаг — своя форма: браузер проверит обязательные поля шага, прежде чем пустить дальше.
  const onStepSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (step < STEPS.length) setStep(step + 1);
    else setIsSubmitted(true);
  };

  if (isSubmitted) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-16 text-center space-y-6">
        <div className="w-16 h-16 bg-emerald-100 text-emerald-600 rounded-full flex items-center justify-center mx-auto">
          <CheckCircle2 className="w-10 h-10" />
        </div>
        <h1 className="text-3xl font-bold text-slate-900">Материал успешно отправлен на модерацию</h1>
        <p className="text-slate-600 text-base leading-relaxed">
          Спасибо за ваш вклад в сохранение культурного наследия Астрахани! Историки и архивисты Центра проверят
          предоставленные данные в течение 1–3 рабочих дней.
        </p>
        <div className="pt-4 flex justify-center">
          <Link
            href="/map"
            className="inline-flex items-center gap-2 px-6 py-3 bg-slate-900 text-white rounded-lg font-semibold text-sm hover:bg-slate-800 transition-colors"
          >
            Перейти к интерактивной карте
          </Link>
        </div>
      </div>
    );
  }

  const back = (
    <button type="button" onClick={() => setStep(step - 1)} className={backClass}>
      <ArrowLeft className="w-4 h-4" />
      <span>Назад</span>
    </button>
  );

  return (
    <div className="max-w-3xl mx-auto px-4 sm:px-6 py-10 sm:py-14 w-full">
      <PageHeader
        className="mb-6"
        kicker="Краудсорсинг архива"
        title="Предложить объект или архивный снимок"
        lead="Поделитесь информацией об утраченном или малоизвестном памятнике архитектуры."
      />

      <div className="mb-8 p-4 bg-slate-50 border border-slate-200 rounded-xl flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs">
        <div className="flex items-center gap-2.5 text-slate-700">
          <UserCheck className="w-5 h-5 text-blue-600 shrink-0" />
          <span>Вы можете подать заявку как зарегистрированный участник, чтобы закрепить авторство в реестре.</span>
        </div>
        <Link
          href="/login"
          className="px-3.5 py-1.5 bg-white hover:bg-slate-100 border border-slate-200 rounded-lg font-semibold text-slate-800 transition-colors shrink-0 flex items-center justify-center gap-1.5"
        >
          <LogIn className="w-3.5 h-3.5" />
          <span>Войти в профиль</span>
        </Link>
      </div>

      <Stepper step={step} />

      <form key={step} onSubmit={onStepSubmit} className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-xs space-y-4">
        {step === 1 && (
          <>
            <div>
              <Label htmlFor={`${id}-title`}>Название объекта (историческое или народное)*</Label>
              <DraftInput id={`${id}-title`} required autoFocus placeholder="Например: Дом купца Франгулова или Усадьба на Набережной" value={draft.title} onChange={(e) => set({ title: e.target.value })} />
            </div>
            <div>
              <Label htmlFor={`${id}-address`}>Адрес или ориентир в Астрахани*</Label>
              <DraftInput id={`${id}-address`} required placeholder="г. Астрахань, ул. Советская, д. ..." value={draft.address} onChange={(e) => set({ address: e.target.value })} />
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <Label htmlFor={`${id}-category`}>Категория</Label>
                <DraftSelect id={`${id}-category`} value={draft.category} onChange={(e) => set({ category: e.target.value as ObjectDraft['category'] })}>
                  {CATEGORY_OPTIONS.map((o) => (
                    <option key={o.value} value={o.value}>
                      {o.label}
                    </option>
                  ))}
                </DraftSelect>
              </div>
              <div>
                <Label htmlFor={`${id}-year`}>Примерный год постройки</Label>
                <DraftInput id={`${id}-year`} placeholder="Например: 1894 г. или конец XIX века" value={draft.yearBuilt} onChange={(e) => set({ yearBuilt: e.target.value })} />
              </div>
            </div>
            <div className="pt-4 flex justify-end">
              <button type="submit" className={nextClass}>
                <span>Далее: История</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </>
        )}

        {step === 2 && (
          <>
            <div>
              <Label htmlFor={`${id}-description`}>Исторические сведения и описание*</Label>
              <DraftTextarea id={`${id}-description`} rows={5} required autoFocus placeholder="Опишите историю здания, известных владельцев, особенности архитектуры..." value={draft.description} onChange={(e) => set({ description: e.target.value })} />
            </div>
            <div>
              <Label htmlFor={`${id}-sources`}>Источники информации (архив, семейный архив, литература)</Label>
              <DraftInput id={`${id}-sources`} placeholder="Например: ГАОО, фонд 12 или устные воспоминания" value={draft.sources} onChange={(e) => set({ sources: e.target.value })} />
            </div>
            <div className="pt-4 flex justify-between">
              {back}
              <button type="submit" className={nextClass}>
                <span>Далее: Загрузка медиа</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </>
        )}

        {step === 3 && (
          <>
            <div>
              <span className="block text-sm font-semibold text-slate-900 mb-1.5">Загрузка фотографий и архивных сканов</span>
              <div className="border-2 border-dashed border-slate-200 rounded-xl p-8 text-center bg-slate-50 hover:bg-slate-100 transition-colors cursor-pointer">
                <Upload className="w-8 h-8 text-slate-400 mx-auto mb-2" />
                <p className="text-sm font-semibold text-slate-700">Нажмите для выбора файлов или перетащите их сюда</p>
                <p className="text-xs text-slate-400 mt-1">Поддерживаются JPG, PNG, WebP (до 25 МБ)</p>
              </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <Label htmlFor={`${id}-name`}>Ваше имя (для указания авторства)*</Label>
                <DraftInput id={`${id}-name`} required placeholder="Иван Петров" value={draft.authorName} onChange={(e) => set({ authorName: e.target.value })} />
              </div>
              <div>
                <Label htmlFor={`${id}-email`}>Контактный Email (для обратной связи)*</Label>
                <DraftInput id={`${id}-email`} type="email" required placeholder="Ваш e-mail" value={draft.authorEmail} onChange={(e) => set({ authorEmail: e.target.value })} />
              </div>
            </div>
            <div className="p-3 bg-slate-50 rounded-lg text-xs text-slate-500 flex items-start gap-2">
              <Info className="w-4 h-4 text-blue-600 shrink-0 mt-0.5" />
              <span>
                Отправляя данные, вы подтверждаете согласие на их научную проверку и публикацию в открытом архиве
                платформы с сохранением вашего авторства.
              </span>
            </div>
            <div className="pt-4 flex justify-between">
              {back}
              <button type="submit" className={submitClass}>
                <CheckCircle2 className="w-4 h-4" />
                <span>Отправить на модерацию</span>
              </button>
            </div>
          </>
        )}
      </form>
    </div>
  );
}
