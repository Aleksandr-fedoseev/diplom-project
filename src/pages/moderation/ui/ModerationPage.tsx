'use client';

import { Suspense } from 'react';
import { RequireRole } from '@/entities/session';
import { roleAtLeast, type User } from '@/entities/user';
import { useTabParam } from '@/shared/lib';
import { PageHeader, Tabs } from '@/shared/ui';
import { ObjectsManager } from './ObjectsManager';
import { SubmissionQueue } from './SubmissionQueue';

const TABS = ['submissions', 'objects'] as const;

function Moderation({ user }: { user: User }) {
  const [tab, setTab] = useTabParam(TABS, 'submissions');
  return (
    <div className="max-w-5xl mx-auto px-4 sm:px-6 py-10 sm:py-14 w-full space-y-6">
      <PageHeader kicker="Модерация" title="Заявки и каталог" lead="Очередь заявок пользователей и редактирование объектов каталога. Автора можно заблокировать при отказе по заявке." />
      <Tabs
        value={tab}
        onChange={setTab}
        items={[
          { id: 'submissions', label: 'Заявки' },
          { id: 'objects', label: 'Объекты' },
        ]}
      />
      {tab === 'submissions' && <SubmissionQueue />}
      {tab === 'objects' && <ObjectsManager canDelete={roleAtLeast(user.role, 'admin')} />}
    </div>
  );
}

export default function ModerationPage() {
  return (
    <Suspense>
      <RequireRole role="moderator">{(user) => <Moderation user={user} />}</RequireRole>
    </Suspense>
  );
}
