import type { OrderStatus } from '../model/types';

export const ORDER_STATUS_LABELS: Record<OrderStatus, string> = {
  pending: 'Ждёт оплаты',
  paid: 'Оплачен',
  shipped: 'Отправлен',
  completed: 'Выполнен',
  cancelled: 'Отменён',
};

// Как на бэкенде (shop.transitions): куда администратор может перевести заказ.
// Оплату подтверждает только ЮKassa, поэтому pending → paid здесь нет.
const NEXT: Partial<Record<OrderStatus, OrderStatus[]>> = {
  pending: ['cancelled'],
  paid: ['shipped', 'cancelled'],
  shipped: ['completed'],
};

export const nextOrderStatuses = (s: OrderStatus): OrderStatus[] => NEXT[s] ?? [];
