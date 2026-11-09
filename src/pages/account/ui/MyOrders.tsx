'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import { ORDER_STATUS_LABELS, orderApi, orderKeys, type Order } from '@/entities/order';
import { productKeys } from '@/entities/product';
import { formatDateTime, formatRub } from '@/shared/lib';
import { ErrorMessage, QueryState } from '@/shared/ui';
import { StatusBadge } from './StatusBadge';

const TONE = { pending: 'amber', paid: 'blue', shipped: 'blue', completed: 'green', cancelled: 'gray' } as const;

function OrderRow({ order }: { order: Order }) {
  const qc = useQueryClient();
  const pay = useMutation({
    mutationFn: () => orderApi.pay(order.id),
    onSuccess: (o) => {
      if (o.paymentUrl) window.location.href = o.paymentUrl;
    },
  });
  const cancel = useMutation({
    mutationFn: () => orderApi.cancel(order.id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: orderKeys.mine });
      qc.invalidateQueries({ queryKey: productKeys.all }); // товары вернулись на склад
    },
  });

  return (
    <li className="p-4 bg-white border border-slate-200 rounded-xl space-y-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <StatusBadge tone={TONE[order.status]}>{ORDER_STATUS_LABELS[order.status]}</StatusBadge>
          <span className="text-xs text-slate-500">
            Заказ № {order.id} · {formatDateTime(order.createdAt)}
          </span>
        </div>
        <span className="font-mono font-bold text-sm">{formatRub(order.totalKopecks / 100)}</span>
      </div>
      <ul className="text-xs text-slate-600">
        {order.items.map((i) => (
          <li key={i.productId}>
            {i.title} × {i.quantity}
          </li>
        ))}
      </ul>
      {order.status === 'pending' && (
        <div className="flex gap-2 pt-1">
          <button
            type="button"
            disabled={pay.isPending}
            onClick={() => pay.mutate()}
            className="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-xs font-bold"
          >
            Оплатить
          </button>
          <button
            type="button"
            disabled={cancel.isPending}
            onClick={() => confirm(`Отменить заказ № ${order.id}?`) && cancel.mutate()}
            className="px-3 py-1.5 bg-slate-100 hover:bg-slate-200 rounded-lg text-xs font-bold"
          >
            Отменить
          </button>
        </div>
      )}
      {(pay.error || cancel.error) && <ErrorMessage error={pay.error ?? cancel.error} />}
    </li>
  );
}

export function MyOrders() {
  const query = useQuery({ queryKey: orderKeys.mine, queryFn: orderApi.mine });
  return (
    <QueryState query={query}>
      {(list) =>
        list.length === 0 ? (
          <div className="py-12 text-center text-sm text-slate-500 space-y-3">
            <p>Заказов пока нет.</p>
            <Link href="/shop" className="text-blue-600 font-semibold">
              В магазин →
            </Link>
          </div>
        ) : (
          <ul className="space-y-3">
            {list.map((o) => (
              <OrderRow key={o.id} order={o} />
            ))}
          </ul>
        )
      }
    </QueryState>
  );
}
