'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { nextOrderStatuses, ORDER_STATUS_LABELS, orderApi, type OrderStatus } from '@/entities/order';
import { formatDateTime, formatRub } from '@/shared/lib';
import { Chip, ErrorMessage, QueryState } from '@/shared/ui';

const FILTERS: (OrderStatus | '')[] = ['', 'pending', 'paid', 'shipped', 'completed', 'cancelled'];

// Заказы всех покупателей. Кнопки — только допустимые переходы статуса (как на сервере).
export function OrdersAdmin() {
  const qc = useQueryClient();
  const [status, setStatus] = useState<OrderStatus | ''>('');
  const orders = useQuery({ queryKey: ['orders', 'admin', status], queryFn: () => orderApi.adminList(status) });
  const change = useMutation({
    mutationFn: (v: { id: number; status: OrderStatus }) => orderApi.setStatus(v.id, v.status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['orders'] }),
  });

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-2">
        {FILTERS.map((s) => (
          <Chip key={s || 'all'} active={status === s} onClick={() => setStatus(s)} className="px-3 py-1.5 text-xs">
            {s ? ORDER_STATUS_LABELS[s] : 'Все'}
          </Chip>
        ))}
      </div>
      {change.error && <ErrorMessage error={change.error} />}
      <QueryState query={orders}>
        {(list) =>
          list.length === 0 ? (
            <p className="py-12 text-center text-sm text-slate-500">Заказов нет.</p>
          ) : (
            <ul className="space-y-2">
              {list.map((o) => (
                <li key={o.id} className="p-4 bg-white border border-slate-200 rounded-xl space-y-2">
                  <div className="flex flex-wrap items-center justify-between gap-2 text-sm">
                    <span className="font-bold">
                      № {o.id} · {ORDER_STATUS_LABELS[o.status]}
                    </span>
                    <span className="font-mono font-bold">{formatRub(o.totalKopecks / 100)}</span>
                  </div>
                  <div className="text-xs text-slate-500">
                    {formatDateTime(o.createdAt)} · {o.phone} · {o.deliveryAddress}
                  </div>
                  <div className="text-xs text-slate-700">{o.items.map((i) => `${i.title} × ${i.quantity}`).join(', ')}</div>
                  {nextOrderStatuses(o.status).length > 0 && (
                    <div className="flex gap-2 pt-1">
                      {nextOrderStatuses(o.status).map((next) => (
                        <button
                          key={next}
                          type="button"
                          disabled={change.isPending}
                          onClick={() =>
                            confirm(`Заказ № ${o.id}: «${ORDER_STATUS_LABELS[o.status]}» → «${ORDER_STATUS_LABELS[next]}»?`) &&
                            change.mutate({ id: o.id, status: next })
                          }
                          className={
                            next === 'cancelled'
                              ? 'px-3 py-1 rounded-lg text-xs font-bold bg-rose-100 text-rose-800'
                              : 'px-3 py-1 rounded-lg text-xs font-bold bg-slate-900 text-white'
                          }
                        >
                          {ORDER_STATUS_LABELS[next]}
                        </button>
                      ))}
                    </div>
                  )}
                </li>
              ))}
            </ul>
          )
        }
      </QueryState>
    </div>
  );
}
