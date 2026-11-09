'use client';

import type { ReactNode } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { ShieldAlert } from 'lucide-react';
import { ROLE_LABELS, type Role, type User } from '@/entities/user/@x/session';
import { Loading } from '@/shared/ui';
import { useSession } from '../model/useSession';

interface RequireRoleProps {
  role: Role;
  children: (user: User) => ReactNode;
}

// Страница только для вошедших с ролью не ниже role. Это удобство интерфейса:
// сервер всё равно проверяет права на каждый запрос и ответит 403.
export function RequireRole({ role, children }: RequireRoleProps) {
  const { user, isLoading, hasRole } = useSession();
  const pathname = usePathname() ?? '/';

  if (isLoading) return <Loading />;
  if (!user) {
    return (
      <div className="max-w-md mx-auto px-4 py-16 text-center space-y-4">
        <h1 className="text-2xl font-bold text-slate-900">Нужно войти</h1>
        <Link
          href={`/login?next=${encodeURIComponent(pathname)}`}
          className="inline-flex px-6 py-3 bg-slate-900 text-white rounded-lg font-semibold text-sm hover:bg-slate-800"
        >
          Войти
        </Link>
      </div>
    );
  }
  if (!hasRole(role)) {
    return (
      <div className="max-w-md mx-auto px-4 py-16 text-center space-y-3">
        <ShieldAlert className="w-10 h-10 text-rose-500 mx-auto" />
        <h1 className="text-2xl font-bold text-slate-900">Недостаточно прав</h1>
        <p className="text-sm text-slate-600">
          Раздел доступен роли «{ROLE_LABELS[role]}» и выше. Ваша роль — «{ROLE_LABELS[user.role]}».
        </p>
      </div>
    );
  }
  return <>{children(user)}</>;
}
