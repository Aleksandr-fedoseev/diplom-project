'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { 
  Landmark, 
  Mail, 
  Lock, 
  User, 
  ArrowRight, 
  ShieldCheck, 
  CheckCircle2,
  Info
} from 'lucide-react';

export default function LoginPage() {
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [role, setRole] = useState<'researcher' | 'citizen'>('citizen');

  const handleAuth = (e: React.FormEvent) => {
    e.preventDefault();
    alert(`Демонстрационный режим: успешный ${mode === 'login' ? 'вход' : 'аккаунт создан'} для ${email}!`);
  };

  return (
    <div className="min-h-[calc(100vh-10rem)] flex items-center justify-center px-4 sm:px-6 lg:px-8 py-12 bg-slate-50">
      <div className="max-w-md w-full space-y-6">
        
        {/* Top Header Card */}
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

        {/* Auth Mode Tabs */}
        <div className="bg-slate-200/80 p-1 rounded-xl flex text-xs font-semibold">
          <button
            type="button"
            onClick={() => setMode('login')}
            className={`flex-1 py-2 rounded-lg transition-all ${
              mode === 'login'
                ? 'bg-white text-slate-900 shadow-xs font-bold'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            Вход в аккаунт
          </button>
          <button
            type="button"
            onClick={() => setMode('register')}
            className={`flex-1 py-2 rounded-lg transition-all ${
              mode === 'register'
                ? 'bg-white text-slate-900 shadow-xs font-bold'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            Регистрация
          </button>
        </div>

        {/* Main Form Container */}
        <div className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-xs space-y-5">
          
          <form onSubmit={handleAuth} className="space-y-4">
            
            {mode === 'register' && (
              <div>
                <label className="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                  ФИО / Имя пользователя
                </label>
                <div className="relative">
                  <User className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
                  <input
                    type="text"
                    required
                    placeholder="Александр Иванов"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    className="w-full bg-slate-50 border border-slate-200 pl-10 pr-3 py-2.5 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
                  />
                </div>
              </div>
            )}

            <div>
              <label className="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                Электронная почта (Email)
              </label>
              <div className="relative">
                <Mail className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
                <input
                  type="email"
                  required
                  placeholder="Ваш e-mail"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full bg-slate-50 border border-slate-200 pl-10 pr-3 py-2.5 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
                />
              </div>
            </div>

            <div>
              <div className="flex items-center justify-between mb-1.5">
                <label className="block text-xs font-bold text-slate-700 uppercase tracking-wider">
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
              <div className="relative">
                <Lock className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
                <input
                  type="password"
                  required
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full bg-slate-50 border border-slate-200 pl-10 pr-3 py-2.5 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
                />
              </div>
            </div>

            {mode === 'register' && (
              <div>
                <label className="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                  Статус участника
                </label>
                <select
                  value={role}
                  onChange={(e) => setRole(e.target.value as any)}
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

          {/* Social / External Auth Options */}
          <div className="space-y-3 pt-2">
            <div className="relative flex items-center justify-center">
              <div className="border-t border-slate-200 w-full" />
              <span className="bg-white px-3 text-[11px] font-semibold text-slate-400 uppercase tracking-wider absolute">
                или через сервисы
              </span>
            </div>

            <div className="grid grid-cols-2 gap-2.5 pt-2">
              <button
                type="button"
                onClick={() => alert('Демо: Вход через Яндекс ID')}
                className="py-2.5 px-3 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-lg text-xs font-semibold text-slate-800 transition-colors text-center"
              >
                Яндекс ID
              </button>

              <button
                type="button"
                onClick={() => alert('Демо: Вход через Госуслуги (ЕСИА)')}
                className="py-2.5 px-3 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-lg text-xs font-semibold text-slate-800 transition-colors text-center"
              >
                Госуслуги (ЕСИА)
              </button>
            </div>
          </div>

        </div>

        {/* Info notice */}
        <div className="bg-blue-50 border border-blue-200 rounded-xl p-3.5 text-xs text-blue-900 flex items-start gap-2.5">
          <Info className="w-4 h-4 text-blue-600 shrink-0 mt-0.5" />
          <span>
            Авторизованные пользователи могут предлагать архивные фото, указывать точки утраченных зданий и отслеживать статус научной модерации.
          </span>
        </div>

      </div>
    </div>
  );
}
