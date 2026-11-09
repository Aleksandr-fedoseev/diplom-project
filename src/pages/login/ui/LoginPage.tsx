'use client';

import { Suspense, useState } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { Info, Landmark } from 'lucide-react';
import { useSession } from '@/entities/session';
import { cn } from '@/shared/lib';
import { safeNext } from '../lib/safeNext';
import { AuthForm, type AuthMode } from './AuthForm';

const MODES: { id: AuthMode; label: string }[] = [
  { id: 'login', label: 'Вход в аккаунт' },
  { id: 'register', label: 'Регистрация' },
];

function Login() {
  const [mode, setMode] = useState<AuthMode>('login');
  const next = safeNext(useSearchParams()?.get('next'));
  const { user } = useSession();

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
          {user ? (
            <p className="text-sm text-slate-700 text-center">
              Вы вошли как <b>{user.name}</b>.{' '}
              <Link href={next} className="text-blue-600 font-semibold">
                Продолжить →
              </Link>
            </p>
          ) : (
            <AuthForm key={mode} mode={mode} next={next} />
          )}

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

export default function LoginPage() {
  return (
    <Suspense>
      <Login />
    </Suspense>
  );
}
