'use client';

import { useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { LogOut, User } from 'lucide-react';
import { useSession } from '@/entities/session';
import { ROLE_LABELS, UserAvatar } from '@/entities/user';
import { accountLinks } from '../model/nav';

const iconButton = 'p-2 text-slate-600 hover:text-slate-950 hover:bg-slate-100 rounded-lg transition-colors';

// Гостю — ссылка «Войти», вошедшему — меню профиля и выход. Рабочие разделы сотрудников —
// не здесь, а вкладками в шапке.
export function UserMenu() {
  const { user, logout } = useSession();
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const close = (e: MouseEvent) => !ref.current?.contains(e.target as Node) && setOpen(false);
    const esc = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(false);
    document.addEventListener('mousedown', close);
    document.addEventListener('keydown', esc);
    return () => {
      document.removeEventListener('mousedown', close);
      document.removeEventListener('keydown', esc);
    };
  }, [open]);

  if (!user) {
    return (
      <Link href="/login" className={iconButton} aria-label="Войти">
        <User className="w-5 h-5" />
      </Link>
    );
  }

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        aria-haspopup="menu"
        className="flex items-center gap-2 px-2.5 py-1.5 rounded-lg hover:bg-slate-100 text-sm font-semibold text-slate-800"
      >
        <UserAvatar user={user} className="w-7 h-7 text-xs" />
        <span className="hidden md:inline max-w-[140px] truncate">{user.name}</span>
      </button>
      {open && (
        <div role="menu" className="absolute right-0 mt-2 w-60 bg-white border border-slate-200 rounded-xl shadow-lg p-2 z-50">
          <div className="px-3 py-2 border-b border-slate-100 mb-1">
            <div className="text-sm font-bold text-slate-900 truncate">{user.name}</div>
            <div className="text-xs text-slate-500 truncate">{user.email}</div>
            <div className="text-[10px] font-bold uppercase tracking-wider text-blue-600 mt-1">{ROLE_LABELS[user.role]}</div>
          </div>
          {accountLinks(user.role).map((l) => (
            <Link
              key={l.href}
              href={l.href}
              role="menuitem"
              onClick={() => setOpen(false)}
              className="block px-3 py-2 rounded-lg text-sm text-slate-700 hover:bg-slate-100"
            >
              {l.label}
            </Link>
          ))}
          <button
            type="button"
            role="menuitem"
            onClick={() => {
              setOpen(false);
              logout();
              router.push('/');
            }}
            className="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-rose-700 hover:bg-rose-50"
          >
            <LogOut className="w-4 h-4" />
            <span>Выйти</span>
          </button>
        </div>
      )}
    </div>
  );
}
