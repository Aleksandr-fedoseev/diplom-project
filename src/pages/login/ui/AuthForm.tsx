'use client';

import React, { useId, useState } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowRight, Lock, Mail, User } from 'lucide-react';
import { sessionApi, useSession } from '@/entities/session';
import { ErrorMessage } from '@/shared/ui';
import { IconInput } from './IconInput';

export type AuthMode = 'login' | 'register';

const label = 'block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5';

// Вход и регистрация через /api/auth. После регистрации сразу выполняется вход.
export function AuthForm({ mode, next }: { mode: AuthMode; next: string }) {
  const id = useId();
  const router = useRouter();
  const { login } = useSession();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<unknown>(null);
  const [pending, setPending] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setPending(true);
    try {
      if (mode === 'register') await sessionApi.register(name.trim(), email.trim(), password);
      await login(email.trim(), password);
      router.push(next);
    } catch (err) {
      setError(err);
      setPending(false);
    }
  };

  return (
    <form onSubmit={onSubmit} className="space-y-4">
      {mode === 'register' && (
        <div>
          <label htmlFor={`${id}-name`} className={label}>
            Имя
          </label>
          <IconInput id={`${id}-name`} icon={User} required autoComplete="name" placeholder="Александр Иванов" value={name} onChange={(e) => setName(e.target.value)} />
        </div>
      )}

      <div>
        <label htmlFor={`${id}-email`} className={label}>
          Электронная почта
        </label>
        <IconInput id={`${id}-email`} icon={Mail} type="email" required autoComplete="email" placeholder="Ваш e-mail" value={email} onChange={(e) => setEmail(e.target.value)} />
      </div>

      <div>
        <label htmlFor={`${id}-password`} className={label}>
          Пароль
        </label>
        <IconInput
          id={`${id}-password`}
          icon={Lock}
          type="password"
          required
          // Как на сервере: не короче 8 символов и не длиннее 72 байт (предел bcrypt).
          minLength={mode === 'register' ? 8 : undefined}
          maxLength={72}
          autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
          placeholder="••••••••"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        {mode === 'register' && <p className="text-[11px] text-slate-400 mt-1">Не меньше 8 символов</p>}
      </div>

      {error ? <ErrorMessage error={error} /> : null}

      <button
        type="submit"
        disabled={pending}
        className="w-full py-3 px-4 bg-slate-900 hover:bg-blue-600 disabled:bg-slate-400 text-white rounded-lg text-sm font-semibold shadow-xs transition-colors flex items-center justify-center gap-2 mt-2"
      >
        <span>{pending ? 'Подождите...' : mode === 'login' ? 'Войти' : 'Зарегистрироваться'}</span>
        <ArrowRight className="w-4 h-4" />
      </button>
    </form>
  );
}
