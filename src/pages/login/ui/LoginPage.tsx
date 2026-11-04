'use client';

import { useState } from 'react';
import { Info, Landmark } from 'lucide-react';
import { cn } from '@/shared/lib';
import { AuthForm, type AuthMode } from './AuthForm';

const MODES: { id: AuthMode; label: string }[] = [
  { id: 'login', label: 'Вход в аккаунт' },
  { id: 'register', label: 'Регистрация' },
];

const SERVICES = ['Яндекс ID', 'Госуслуги (ЕСИА)'];

export default function LoginPage() {
  const [mode, setMode] = useState<AuthMode>('login');

  return (
    <div className="min-h-[calc(100vh-10rem)] flex items-center justify-center px-4 sm:px-6 lg:px-8 py-12 bg-slate-50">
      <div className="max-w-md w-full space-y-6">
        <div className="text-center space-y-2">
          <div className="w-12 h-12 rounded-xl bg-slate-900 flex items-center justify-center text-white mx-auto shadow-sm">
            <Landmark className="w-6 h-6" />
          </div>
          <h1 className="text-2xl sm:text-3xl font-bold text-slate-900 tracking-tight">
            {mode === 'login' ? 'Авторизация в системе' : 'Регистрация исследователя'}
          </h1>
          <p className="text-xs sm:text-sm text-slate-500 max-w-sm mx-auto">
            Единый доступ к предложениям объектов, загрузке архивных материалов и истории модерации
          </p>
        </div>

        <div role="tablist" className="bg-slate-200/80 p-1 rounded-xl flex text-xs font-semibold">
          {MODES.map((m) => (
            <button
              key={m.id}
              type="button"
              role="tab"
              aria-selected={mode === m.id}
              onClick={() => setMode(m.id)}
              className={cn(
                'flex-1 py-2 rounded-lg transition-all',
                mode === m.id ? 'bg-white text-slate-900 shadow-xs font-bold' : 'text-slate-600 hover:text-slate-900',
              )}
            >
              {m.label}
            </button>
          ))}
        </div>

        <div className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-xs space-y-5">
          {/* key: при смене режима поля начинаются заново */}
          <AuthForm key={mode} mode={mode} />

          <div className="space-y-3 pt-2">
            <div className="relative flex items-center justify-center">
              <div className="border-t border-slate-200 w-full" />
              <span className="bg-white px-3 text-[11px] font-semibold text-slate-400 uppercase tracking-wider absolute">
                или через сервисы
              </span>
            </div>
            <div className="grid grid-cols-2 gap-2.5 pt-2">
              {SERVICES.map((service) => (
                <button
                  key={service}
                  type="button"
                  onClick={() => alert(`Демо: вход через ${service}`)}
                  className="py-2.5 px-3 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-lg text-xs font-semibold text-slate-800 transition-colors text-center"
                >
                  {service}
                </button>
              ))}
            </div>
          </div>
        </div>

        <div className="bg-blue-50 border border-blue-200 rounded-xl p-3.5 text-xs text-blue-900 flex items-start gap-2.5">
          <Info className="w-4 h-4 text-blue-600 shrink-0 mt-0.5" />
          <span>
            Авторизованные пользователи могут предлагать архивные фото, указывать точки утраченных зданий и отслеживать
            статус научной модерации.
          </span>
        </div>
      </div>
    </div>
  );
}
