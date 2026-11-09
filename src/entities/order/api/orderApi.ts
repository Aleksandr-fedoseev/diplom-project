import { api } from '@/shared/api';
import type { Order, OrderStatus } from '../model/types';

export const orderKeys = {
  mine: ['orders', 'mine'] as const,
  admin: (status: string) => ['orders', 'admin', status] as const,
};

export const orderApi = {
  // Заказ из текущей корзины; корзина на сервере очищается.
  checkout: (deliveryAddress: string, phone: string) =>
    api<Order>('/orders', { method: 'POST', json: { deliveryAddress, phone } }),
  mine: () => api<Order[]>('/me/orders'),
  pay: (id: number) => api<Order>(`/orders/${id}/pay`, { method: 'POST' }),
  cancel: (id: number) => api<Order>(`/orders/${id}/cancel`, { method: 'POST' }),
  adminList: (status = '') => api<Order[]>('/admin/orders', { query: { status, limit: 200 } }),
  setStatus: (id: number, status: OrderStatus) =>
    api<Order>(`/admin/orders/${id}/status`, { method: 'PUT', json: { status } }),
};
