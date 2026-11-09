'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useToken } from '@/shared/api';
import { cartApi, cartKeys } from '../api/cartApi';
import type { Cart } from './types';

// Корзина вошедшего пользователя. Каждое изменение возвращает с сервера новую корзину.
export function useCart() {
  const token = useToken();
  const qc = useQueryClient();
  const cart = useQuery({ queryKey: cartKeys.cart, queryFn: cartApi.get, enabled: !!token });
  const onSuccess = (c: Cart) => qc.setQueryData(cartKeys.cart, c);

  const setItem = useMutation({
    mutationFn: ({ productId, quantity }: { productId: number; quantity: number }) => cartApi.setItem(productId, quantity),
    onSuccess,
  });
  const removeItem = useMutation({ mutationFn: cartApi.removeItem, onSuccess });

  const items = cart.data?.items ?? [];
  const quantityOf = (productId: number) => items.find((i) => i.productId === productId)?.quantity ?? 0;

  return {
    cart: cart.data,
    isLoading: cart.isPending && !!token,
    count: items.reduce((n, i) => n + i.quantity, 0),
    add: (productId: number) => setItem.mutateAsync({ productId, quantity: quantityOf(productId) + 1 }),
    setQuantity: (productId: number, quantity: number) => setItem.mutateAsync({ productId, quantity }),
    remove: (productId: number) => removeItem.mutateAsync(productId),
    isUpdating: setItem.isPending || removeItem.isPending,
    error: setItem.error ?? removeItem.error ?? cart.error,
  };
}
