'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowDown, ArrowUp, X } from 'lucide-react';
import { useObjects } from '@/entities/heritage-object';
import { formatDistance, formatDuration, routeApi, routeKeys } from '@/entities/route';
import { slugify } from '@/shared/lib';
import { ErrorMessage, Field, FormSelect, Input, QueryState, Textarea } from '@/shared/ui';

// Создание тематического маршрута: объекты и порядок посещения; линию по улицам строит OSRM.
function RouteForm({ onDone }: { onDone: () => void }) {
  const qc = useQueryClient();
  const objects = useObjects();
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [stops, setStops] = useState<string[]>([]);
  const titleOf = (slug: string) => objects.data?.find((o) => o.slug === slug)?.title ?? slug;
  const move = (i: number, d: -1 | 1) =>
    setStops((s) => {
      const next = [...s];
      [next[i], next[i + d]] = [next[i + d], next[i]];
      return next;
    });

  const create = useMutation({
    mutationFn: () => {
      if (stops.length < 2) throw new Error('В маршруте нужно хотя бы два объекта');
      return routeApi.createThematic({ slug: slugify(title), title: title.trim(), description: description.trim(), objects: stops });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: routeKeys.thematic });
      onDone();
    },
  });

  return (
    <form
      className="p-5 bg-white border border-blue-200 rounded-xl space-y-3"
      onSubmit={(e) => {
        e.preventDefault();
        create.mutate();
      }}
    >
      <Field label="Название" required>
        <Input required value={title} onChange={(e) => setTitle(e.target.value)} />
      </Field>
      <Field label="Описание">
        <Textarea rows={2} value={description} onChange={(e) => setDescription(e.target.value)} />
      </Field>
      <Field label="Добавить объект">
        <FormSelect value="" onChange={(e) => e.target.value && setStops((s) => [...s, e.target.value])}>
          <option value="">— выберите объект —</option>
          {(objects.data ?? [])
            .filter((o) => !stops.includes(o.slug))
            .map((o) => (
              <option key={o.slug} value={o.slug}>
                {o.title}
              </option>
            ))}
        </FormSelect>
      </Field>
      {stops.length > 0 && (
        <ol className="space-y-1">
          {stops.map((slug, i) => (
            <li key={slug} className="flex items-center gap-2 p-2 bg-slate-50 rounded-lg text-sm">
              <span className="w-6 h-6 rounded-full bg-slate-900 text-white text-xs font-bold flex items-center justify-center">{i + 1}</span>
              <span className="flex-1 truncate">{titleOf(slug)}</span>
              <button type="button" aria-label="Выше" disabled={i === 0} onClick={() => move(i, -1)} className="p-1 disabled:opacity-30">
                <ArrowUp className="w-4 h-4" />
              </button>
              <button type="button" aria-label="Ниже" disabled={i === stops.length - 1} onClick={() => move(i, 1)} className="p-1 disabled:opacity-30">
                <ArrowDown className="w-4 h-4" />
              </button>
              <button type="button" aria-label="Убрать" onClick={() => setStops((s) => s.filter((x) => x !== slug))} className="p-1 text-rose-600">
                <X className="w-4 h-4" />
              </button>
            </li>
          ))}
        </ol>
      )}
      {create.error && <ErrorMessage error={create.error} />}
      <div className="flex gap-2">
        <button type="submit" disabled={create.isPending} className="px-4 py-2 bg-slate-900 text-white rounded-lg text-xs font-bold uppercase">
          {create.isPending ? 'Строим маршрут...' : 'Создать маршрут'}
        </button>
        <button type="button" onClick={onDone} className="px-4 py-2 bg-slate-200 rounded-lg text-xs font-bold uppercase">
          Отмена
        </button>
      </div>
    </form>
  );
}

// Управление тематическими маршрутами: создание и удаление. Изменить маршрут — удалить
// и создать заново: отдельного изменения в API пока нет.
export function RoutesAdmin() {
  const qc = useQueryClient();
  const routes = useQuery({ queryKey: routeKeys.thematic, queryFn: routeApi.thematic });
  const [creating, setCreating] = useState(false);
  const remove = useMutation({
    mutationFn: routeApi.remove,
    onSuccess: () => qc.invalidateQueries({ queryKey: routeKeys.thematic }),
  });

  return (
    <div className="space-y-4">
      {creating ? (
        <RouteForm onDone={() => setCreating(false)} />
      ) : (
        <button type="button" onClick={() => setCreating(true)} className="px-4 py-2.5 bg-slate-900 text-white rounded-xl text-xs font-bold uppercase tracking-wider">
          Новый маршрут
        </button>
      )}
      {remove.error && <ErrorMessage error={remove.error} />}
      <QueryState query={routes}>
        {(list) => (
          <ul className="space-y-2">
            {list.map((r) => (
              <li key={r.id} className="p-4 bg-white border border-slate-200 rounded-xl flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <h3 className="text-sm font-bold text-slate-900">{r.title}</h3>
                  <p className="text-xs text-slate-500">
                    {formatDistance(r.distanceM)} · {formatDuration(r.durationS)} · {r.stops.map((s) => s.title).join(' → ')}
                  </p>
                </div>
                <button
                  type="button"
                  onClick={() => confirm(`Удалить маршрут «${r.title}»?`) && remove.mutate(r.id)}
                  className="text-xs font-bold text-rose-700 hover:text-rose-900 shrink-0"
                >
                  Удалить
                </button>
              </li>
            ))}
          </ul>
        )}
      </QueryState>
    </div>
  );
}
