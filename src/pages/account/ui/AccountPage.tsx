'use client';

import { Suspense } from 'react';
import { RequireRole } from '@/entities/session';
import { isStaff, ROLE_LABELS, UserAvatar, type User } from '@/entities/user';
import { useTabParam } from '@/shared/lib';
import { Tabs } from '@/shared/ui';
import { MyOrders } from './MyOrders';
import { MyRoutes } from './MyRoutes';
import { MySubmissions } from './MySubmissions';
import { Profile } from './Profile';

const TABS = ['profile', 'submissions', 'orders', 'routes'] as const;
type Tab = (typeof TABS)[number];

const TAB_LABELS: Record<Tab, string> = {
  profile: 'Профиль',
  submissions: 'Мои заявки',
  orders: 'Мои заказы',
  routes: 'Мои маршруты',
};

// Сотруднику — профиль и маршруты; заявки и заказы — разделы обычного пользователя.
const tabsFor = (user: User): readonly Tab[] => (isStaff(user.role) ? ['profile', 'routes'] : TABS);

function Account({ user }: { user: User }) {
  const tabs = tabsFor(user);
  const [tab, setTab] = useTabParam(tabs, 'profile');
  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 py-10 sm:py-14 w-full space-y-6">
      <div className="flex items-center gap-4">
        <UserAvatar user={user} className="w-14 h-14 text-xl" />
        <div>
          <div className="text-xs font-bold uppercase tracking-wider text-blue-600">Личный кабинет · {ROLE_LABELS[user.role]}</div>
          <h1 className="text-2xl sm:text-3xl font-bold text-slate-900">{user.name}</h1>
        </div>
      </div>
      <Tabs
        value={tab}
        onChange={setTab}
        items={tabs.map((id) => ({ id, label: TAB_LABELS[id] }))}
      />
      {tab === 'profile' && <Profile user={user} />}
      {tab === 'submissions' && <MySubmissions />}
      {tab === 'orders' && <MyOrders />}
      {tab === 'routes' && <MyRoutes />}
    </div>
  );
}

export default function AccountPage() {
  return (
    <Suspense>
      <RequireRole role="user">{(user) => <Account user={user} />}</RequireRole>
    </Suspense>
  );
}
