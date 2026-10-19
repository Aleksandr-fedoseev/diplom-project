'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { 
  MapPin, 
  FileText, 
  Upload, 
  CheckCircle2, 
  ArrowRight, 
  ArrowLeft,
  ShieldCheck,
  Info,
  UserCheck,
  LogIn
} from 'lucide-react';

export default function ContributePage() {
  const [step, setStep] = useState<number>(1);
  const [formData, setFormData] = useState({
    address: '',
    title: '',
    category: 'wooden',
    yearBuilt: '',
    status: 'identified',
    description: '',
    sources: '',
    authorName: '',
    authorEmail: '',
  });
  const [isSubmitted, setIsSubmitted] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitted(true);
  };

  if (isSubmitted) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-16 text-center space-y-6">
        <div className="w-16 h-16 bg-emerald-100 text-emerald-600 rounded-full flex items-center justify-center mx-auto">
          <CheckCircle2 className="w-10 h-10" />
        </div>
        <h1 className="text-3xl font-bold text-slate-900">
          Материал успешно отправлен на модерацию
        </h1>
        <p className="text-slate-600 text-base leading-relaxed">
          Спасибо за ваш вклад в сохранение культурного наследия Астрахани! Историки и архивисты Центра проверят предоставленные данные в течение 1–3 рабочих дней.
        </p>
        <div className="pt-4 flex justify-center gap-3">
          <Link
            href="/map"
            className="inline-flex items-center gap-2 px-6 py-3 bg-slate-900 text-white rounded-lg font-semibold text-sm hover:bg-slate-800 transition-colors"
          >
            <span>Перейти к интерактивной карте</span>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto px-4 sm:px-6 py-10 sm:py-14 w-full">
      
      {/* Title */}
      <div className="mb-6">
        <div className="text-xs font-bold uppercase tracking-wider text-blue-600 mb-1">
          Краудсорсинг архива
        </div>
        <h1 className="text-3xl sm:text-4xl font-bold text-slate-900 tracking-tight">
          Предложить объект или архивный снимок
        </h1>
        <p className="text-base text-slate-600 mt-1">
          Поделитесь информацией об утраченном или малоизвестном памятнике архитектуры.
        </p>
      </div>

      {/* Auth Banner Notice */}
      <div className="mb-8 p-4 bg-slate-50 border border-slate-200 rounded-xl flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs">
        <div className="flex items-center gap-2.5 text-slate-700">
          <UserCheck className="w-5 h-5 text-blue-600 shrink-0" />
          <span>
            Вы можете подать заявку как зарегистрированный участник, чтобы закрепить авторство в реестре.
          </span>
        </div>
        <Link
          href="/login"
          className="px-3.5 py-1.5 bg-white hover:bg-slate-100 border border-slate-200 rounded-lg font-semibold text-slate-800 transition-colors shrink-0 flex items-center justify-center gap-1.5"
        >
          <LogIn className="w-3.5 h-3.5" />
          <span>Войти в профиль</span>
        </Link>
      </div>

      {/* Stepper Header */}
      <div className="grid grid-cols-3 gap-2 mb-8 text-xs font-semibold">
        <div className={`p-3 rounded-lg border text-center transition-colors ${
          step === 1 ? 'bg-blue-50 border-blue-600 text-blue-700 font-bold' : 'bg-white border-slate-200 text-slate-500'
        }`}>
          1. Локация и адрес
        </div>
        <div className={`p-3 rounded-lg border text-center transition-colors ${
          step === 2 ? 'bg-blue-50 border-blue-600 text-blue-700 font-bold' : 'bg-white border-slate-200 text-slate-500'
        }`}>
          2. Описание и история
        </div>
        <div className={`p-3 rounded-lg border text-center transition-colors ${
          step === 3 ? 'bg-blue-50 border-blue-600 text-blue-700 font-bold' : 'bg-white border-slate-200 text-slate-500'
        }`}>
          3. Фотографии и автор
        </div>
      </div>

      {/* Form Container */}
      <form onSubmit={handleSubmit} className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-xs space-y-6">
        
        {/* STEP 1 */}
        {step === 1 && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-semibold text-slate-900 mb-1.5">
                Название объекта (историческое или народное)*
              </label>
              <input
                type="text"
                required
                placeholder="Например: Дом купца Франгулова или Усадьба на Набережной"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                className="w-full bg-slate-50 border border-slate-200 p-3 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
              />
            </div>

            <div>
              <label className="block text-sm font-semibold text-slate-900 mb-1.5">
                Адрес или ориентир в Астрахани*
              </label>
              <input
                type="text"
                required
                placeholder="г. Астрахань, ул. Советская, д. ..."
                value={formData.address}
                onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                className="w-full bg-slate-50 border border-slate-200 p-3 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
              />
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-semibold text-slate-900 mb-1.5">
                  Категория
                </label>
                <select
                  value={formData.category}
                  onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                  className="w-full bg-slate-50 border border-slate-200 p-3 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600"
                >
                  <option value="wooden">Деревянное зодчество</option>
                  <option value="manor">Купеческий особняк</option>
                  <option value="temple">Культовое сооружение</option>
                  <option value="modern">Модерн</option>
                  <option value="lost">Утраченное здание</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-semibold text-slate-900 mb-1.5">
                  Примерный год постройки
                </label>
                <input
                  type="text"
                  placeholder="Например: 1894 г. или конец XIX века"
                  value={formData.yearBuilt}
                  onChange={(e) => setFormData({ ...formData, yearBuilt: e.target.value })}
                  className="w-full bg-slate-50 border border-slate-200 p-3 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
                />
              </div>
            </div>

            <div className="pt-4 flex justify-end">
              <button
                type="button"
                onClick={() => setStep(2)}
                className="px-6 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-semibold flex items-center gap-2"
              >
                <span>Далее: История</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}

        {/* STEP 2 */}
        {step === 2 && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-semibold text-slate-900 mb-1.5">
                Исторические сведения и описание*
              </label>
              <textarea
                rows={5}
                required
                placeholder="Опишите историю здания, известных владельцев, особенности архитектуры..."
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="w-full bg-slate-50 border border-slate-200 p-3 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
              />
            </div>

            <div>
              <label className="block text-sm font-semibold text-slate-900 mb-1.5">
                Источники информации (архив, семейный архив, литература)
              </label>
              <input
                type="text"
                placeholder="Например: ГАОО, фонд 12 или устные воспоминания"
                value={formData.sources}
                onChange={(e) => setFormData({ ...formData, sources: e.target.value })}
                className="w-full bg-slate-50 border border-slate-200 p-3 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
              />
            </div>

            <div className="pt-4 flex justify-between">
              <button
                type="button"
                onClick={() => setStep(1)}
                className="px-5 py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-sm font-semibold flex items-center gap-2"
              >
                <ArrowLeft className="w-4 h-4" />
                <span>Назад</span>
              </button>

              <button
                type="button"
                onClick={() => setStep(3)}
                className="px-6 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-semibold flex items-center gap-2"
              >
                <span>Далее: Загрузка медиа</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}

        {/* STEP 3 */}
        {step === 3 && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-semibold text-slate-900 mb-1.5">
                Загрузка фотографий и архивных сканов
              </label>
              <div className="border-2 border-dashed border-slate-200 rounded-xl p-8 text-center bg-slate-50 hover:bg-slate-100 transition-colors cursor-pointer">
                <Upload className="w-8 h-8 text-slate-400 mx-auto mb-2" />
                <p className="text-sm font-semibold text-slate-700">
                  Нажмите для выбора файлов или перетащите их сюда
                </p>
                <p className="text-xs text-slate-400 mt-1">
                  Поддерживаются JPG, PNG, WebP (до 25 МБ)
                </p>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-semibold text-slate-900 mb-1.5">
                  Ваше имя (для указания авторства)*
                </label>
                <input
                  type="text"
                  required
                  placeholder="Иван Петров"
                  value={formData.authorName}
                  onChange={(e) => setFormData({ ...formData, authorName: e.target.value })}
                  className="w-full bg-slate-50 border border-slate-200 p-3 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
                />
              </div>

              <div>
                <label className="block text-sm font-semibold text-slate-900 mb-1.5">
                  Контактный Email (для обратной связи)*
                </label>
                <input
                  type="email"
                  required
                  placeholder="Ваш e-mail"
                  value={formData.authorEmail}
                  onChange={(e) => setFormData({ ...formData, authorEmail: e.target.value })}
                  className="w-full bg-slate-50 border border-slate-200 p-3 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
                />
              </div>
            </div>

            <div className="p-3 bg-slate-50 rounded-lg text-xs text-slate-500 flex items-start gap-2">
              <Info className="w-4 h-4 text-blue-600 shrink-0 mt-0.5" />
              <span>
                Отправляя данные, вы подтверждаете согласие на их научную проверку и публикацию в открытом архиве платформы с сохранением вашего авторства.
              </span>
            </div>

            <div className="pt-4 flex justify-between">
              <button
                type="button"
                onClick={() => setStep(2)}
                className="px-5 py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-sm font-semibold flex items-center gap-2"
              >
                <ArrowLeft className="w-4 h-4" />
                <span>Назад</span>
              </button>

              <button
                type="submit"
                className="px-6 py-2.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-sm font-semibold flex items-center gap-2"
              >
                <CheckCircle2 className="w-4 h-4" />
                <span>Отправить на модерацию</span>
              </button>
            </div>
          </div>
        )}

      </form>

    </div>
  );
}
