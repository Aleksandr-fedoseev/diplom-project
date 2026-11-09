import { api } from '@/shared/api';
import type { Cart } from '../model/types';

export const cartKeys = { cart: ['cart'] as const };

export const cartApi = {
  get: () => api<Cart>('/cart'),
  // Задаёт количество товара (не прибавляет).
  setItem: (productId: number, quantity: number) =>
    api<Cart>(`/cart/items/${productId}`, { method: 'PUT', json: { quantity } }),
  removeItem: (productId: number) => api<Cart>(`/cart/items/${productId}`, { method: 'DELETE' }),
};
