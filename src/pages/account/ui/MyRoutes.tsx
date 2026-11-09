'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { formatDistance, formatDuration, routeApi, routeKeys } from '@/entities/route';
import { ErrorMessage, QueryState } from '@/shared/ui';

export function MyRoutes() {
  const qc = useQueryClient();
  const query = useQuery({ queryKey: routeKeys.mine, queryFn: routeApi.mine });
  const remove = useMutation({
    mutationFn: routeApi.remove,
    onSuccess: () => qc.invalidateQueries({ queryKey: routeKeys.mine }),
  });
  return (
    <QueryState query={query}>
      {(list) =>
        list.length === 0 ? (
          <p className="py-12 text-center text-sm text-slate-500">Сохранённых маршрутов нет. Соберите свой на странице «Маршруты».</p>
        ) : (
          <ul className="space-y-3">
            {remove.error ? <ErrorMessage error={remove.error} /> : null}
            {list.map((r) => (
              <li key={r.id} className="p-4 bg-white border border-slate-200 rounded-xl flex items-start justify-between gap-3">
                <div>
                  <h3 className="text-sm font-bold text-slate-900">{r.title}</h3>
                  <p className="text-xs text-slate-500">
                    {formatDistance(r.distanceM)} · {formatDuration(r.durationS)} · {r.stops.map((s) => s.title).join(' → ')}
                  </p>
                </div>
                <button
                  type="button"
                  onClick={() => confirm(`Удалить маршрут «${r.title}»?`) && remove.mutate(r.id)}
                  className="text-xs font-bold text-rose-700 hover:text-rose-900"
                >
                  Удалить
                </button>
              </li>
            ))}
          </ul>
        )
      }
    </QueryState>
  );
}
