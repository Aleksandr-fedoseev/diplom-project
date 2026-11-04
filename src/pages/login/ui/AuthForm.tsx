'use client';

import React, { useId, useState } from 'react';
import { ArrowRight, Lock, Mail, User } from 'lucide-react';
import { IconInput } from './IconInput';

export type AuthMode = 'login' | 'register';

const label = 'block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5';

// Вход и регистрация. Пока демонстрационные: подключатся к /api/auth вместе с API.
export function AuthForm({ mode }: { mode: AuthMode }) {
  const id = useId();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState<'researcher' | 'citizen'>('citizen');

  return (
    <form
      onSubmit={(e: React.FormEvent) => {
        e.preventDefault();
        alert(`Демонстрационный режим: ${mode === 'login' ? 'успешный вход' : 'аккаунт создан'} для ${email}!`);
      }}
      className="space-y-4"
    >
      {mode === 'register' && (
        <div>
          <label htmlFor={`${id}-name`} className={label}>
            ФИО / Имя пользователя
          </label>
          <IconInput id={`${id}-name`} icon={User} required autoComplete="name" placeholder="Александр Иванов" value={name} onChange={(e) => setName(e.target.value)} />
        </div>
      )}

      <div>
        <label htmlFor={`${id}-email`} className={label}>
          Электронная почта (Email)
        </label>
        <IconInput id={`${id}-email`} icon={Mail} type="email" required autoComplete="email" placeholder="Ваш e-mail" value={email} onChange={(e) => setEmail(e.target.value)} />
      </div>

      <div>
        <div className="flex items-center justify-between mb-1.5">
          <label htmlFor={`${id}-password`} className="block text-xs font-bold text-slate-700 uppercase tracking-wider">
            Пароль
          </label>
          {mode === 'login' && (
            <button
              type="button"
              onClick={() => alert('Ссылка для сброса пароля отправлена на email')}
              className="text-xs text-blue-600 hover:text-blue-800 font-medium"
            >
              Забыли пароль?
            </button>
          )}
        </div>
        <IconInput
          id={`${id}-password`}
          icon={Lock}
          type="password"
          required
          // На сервере пароль не короче 8 символов — проверяем сразу, а не после отправки.
          minLength={mode === 'register' ? 8 : undefined}
          autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
          placeholder="••••••••"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </div>

      {mode === 'register' && (
        <div>
          <label htmlFor={`${id}-role`} className={label}>
            Статус участника
          </label>
          <select
            id={`${id}-role`}
            value={role}
            onChange={(e) => setRole(e.target.value as typeof role)}
            className="w-full bg-slate-50 border border-slate-200 p-2.5 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600"
          >
            <option value="citizen">Житель / Краевед-любитель</option>
            <option value="researcher">Историк / Архитектор / Эксперт</option>
          </select>
        </div>
      )}

      <button
        type="submit"
        className="w-full py-3 px-4 bg-slate-900 hover:bg-blue-600 text-white rounded-lg text-sm font-semibold shadow-xs transition-colors flex items-center justify-center gap-2 mt-2"
      >
        <span>{mode === 'login' ? 'Войти в профиль' : 'Зарегистрироваться'}</span>
        <ArrowRight className="w-4 h-4" />
      </button>
    </form>
  );
}
