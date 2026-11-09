'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { objectApi, objectKeys, useObject, useObjects, type HeritageObject } from '@/entities/heritage-object';
import { cn } from '@/shared/lib';
import { ErrorMessage, Loading, QueryState, SearchInput, Tabs } from '@/shared/ui';
import { EMPTY_OBJECT_FORM, toObjectForm } from '../model/objectForm';
import { ObjectEditor } from './ObjectEditor';
import { ObjectMedia } from './ObjectMedia';

function ObjectPanel({ item, canDelete, onClose }: { item: HeritageObject; canDelete: boolean; onClose: () => void }) {
  // В списке нет истории и файлов — правим полную карточку.
  const card = useObject(item.slug);
  const qc = useQueryClient();
  const [tab, setTab] = useState<'passport' | 'media'>('passport');
  const remove = useMutation({
    mutationFn: () => objectApi.remove(item.id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: objectKeys.all });
      onClose();
    },
  });

  return (
    <div className="p-5 bg-white border border-blue-200 rounded-xl shadow-sm space-y-4">
      <div className="flex items-center justify-between gap-3">
        <h3 className="text-lg font-bold text-slate-900">{item.title}</h3>
        <div className="flex items-center gap-3 text-xs font-bold">
          <Link href={`/objects/${item.slug}`} className="text-blue-600">
            На сайте →
          </Link>
          <button type="button" onClick={onClose} className="text-slate-500 hover:text-slate-900">
            Свернуть
          </button>
        </div>
      </div>
      <Tabs
        size="sm"
        value={tab}
        onChange={setTab}
        items={[
          { id: 'passport', label: 'Паспорт' },
          { id: 'media', label: `Файлы${card.data?.files ? ` (${card.data.files.length})` : ''}` },
        ]}
      />
      {card.error ? (
        <ErrorMessage error={card.error} onRetry={() => card.refetch()} />
      ) : !card.data ? (
        <Loading />
      ) : tab === 'passport' ? (
        <ObjectEditor initial={toObjectForm(card.data)} objectId={item.id} onSaved={onClose} />
      ) : (
        <ObjectMedia object={card.data} />
      )}
      {canDelete && (
        <div className="pt-3 border-t border-slate-100">
          {remove.error && <ErrorMessage className="mb-2" error={remove.error} />}
          <button
            type="button"
            disabled={remove.isPending}
            onClick={() => confirm(`Удалить объект «${item.title}» вместе с файлами? Это необратимо.`) && remove.mutate()}
            className="text-xs font-bold text-rose-700 hover:text-rose-900"
          >
            Удалить объект (администратор)
          </button>
        </div>
      )}
    </div>
  );
}

// Все объекты, включая неопубликованные: правка паспорта, файлы, создание и удаление.
export function ObjectsManager({ canDelete }: { canDelete: boolean }) {
  const objects = useObjects(true);
  const [query, setQuery] = useState('');
  const [open, setOpen] = useState<number | 'new' | null>(null);

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row gap-3">
        <SearchInput className="flex-1" placeholder="Поиск по названию или адресу" value={query} onChange={setQuery} />
        <button
          type="button"
          onClick={() => setOpen('new')}
          className="px-4 py-2.5 bg-slate-900 hover:bg-slate-800 text-white rounded-xl text-xs font-bold uppercase tracking-wider"
        >
          Новый объект
        </button>
      </div>
      {open === 'new' && (
        <div className="p-5 bg-white border border-blue-200 rounded-xl space-y-3">
          <div className="flex justify-between">
            <h3 className="text-lg font-bold">Новый объект</h3>
            <button type="button" onClick={() => setOpen(null)} className="text-xs font-bold text-slate-500">
              Отмена
            </button>
          </div>
          <ObjectEditor initial={EMPTY_OBJECT_FORM} onSaved={() => setOpen(null)} />
        </div>
      )}
      <QueryState query={objects}>
        {(list) => {
          const q = query.trim().toLowerCase();
          const shown = list.filter((o) => !q || o.title.toLowerCase().includes(q) || o.address.toLowerCase().includes(q));
          return (
            <ul className="space-y-2">
              {shown.map((o) =>
                open === o.id ? (
                  <li key={o.id}>
                    <ObjectPanel item={o} canDelete={canDelete} onClose={() => setOpen(null)} />
                  </li>
                ) : (
                  <li key={o.id}>
                    <button
                      type="button"
                      onClick={() => setOpen(o.id)}
                      className="w-full text-left p-3 bg-white border border-slate-200 rounded-xl hover:border-slate-400 flex items-center justify-between gap-3"
                    >
                      <div className="min-w-0">
                        <div className="text-sm font-bold text-slate-900 truncate">{o.title}</div>
                        <div className="text-xs text-slate-500 truncate">{o.address}</div>
                      </div>
                      <div className="flex items-center gap-2 shrink-0 text-[11px] font-bold">
                        {!o.coverImage && <span className="text-amber-700">без фото</span>}
                        <span className={cn('px-2 py-0.5 rounded', o.isPublished ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-200 text-slate-700')}>
                          {o.isPublished ? 'опубликован' : 'черновик'}
                        </span>
                      </div>
                    </button>
                  </li>
                ),
              )}
            </ul>
          );
        }}
      </QueryState>
    </div>
  );
}
