'use client';

import { Suspense } from 'react';
import { RequireRole } from '@/entities/session';
import type { User } from '@/entities/user';
import { useTabParam } from '@/shared/lib';
import { PageHeader, Tabs } from '@/shared/ui';
import { CategoriesAdmin } from './CategoriesAdmin';
import { OrdersAdmin } from './OrdersAdmin';
import { ProductsAdmin } from './ProductsAdmin';
import { RoutesAdmin } from './RoutesAdmin';
import { UsersTable } from './UsersTable';

const TABS = ['users', 'categories', 'routes', 'products', 'orders'] as const;

function Admin({ user }: { user: User }) {
  const [tab, setTab] = useTabParam(TABS, 'users');
  return (
    <div className="max-w-5xl mx-auto px-4 sm:px-6 py-10 sm:py-14 w-full space-y-6">
      <PageHeader kicker="Админ-панель" title="Управление платформой" lead="Пользователи и роли, справочник категорий, тематические маршруты, магазин. Заявки и объекты — во вкладке «Модерация»." />
      <Tabs
        value={tab}
        onChange={setTab}
        items={[
          { id: 'users', label: 'Пользователи и роли' },
          { id: 'categories', label: 'Категории' },
          { id: 'routes', label: 'Маршруты' },
          { id: 'products', label: 'Товары' },
          { id: 'orders', label: 'Заказы' },
        ]}
      />
      {tab === 'users' && <UsersTable actor={user} />}
      {tab === 'categories' && <CategoriesAdmin />}
      {tab === 'routes' && <RoutesAdmin />}
      {tab === 'products' && <ProductsAdmin />}
      {tab === 'orders' && <OrdersAdmin />}
    </div>
  );
}

export default function AdminPage() {
  return (
    <Suspense>
      <RequireRole role="admin">{(user) => <Admin user={user} />}</RequireRole>
    </Suspense>
  );
}
