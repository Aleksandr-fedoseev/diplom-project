'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { CheckCircle2, Minus, Plus, ShoppingBag, Trash2, X } from 'lucide-react';
import { cartKeys, useCart } from '@/entities/cart';
import { orderApi, orderKeys, type Order } from '@/entities/order';
import { productKeys } from '@/entities/product';
import { formatRub } from '@/shared/lib';
import { Button, ErrorMessage, Field, Input, Loading } from '@/shared/ui';

const rub = (kopecks: number) => formatRub(kopecks / 100);

// Корзина, выезжающая справа. Закрывается крестиком, кликом по фону и клавишей Esc.
export function CartDrawer({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient();
  const { cart, isLoading, setQuantity, remove, isUpdating, error } = useCart();
  const [address, setAddress] = useState('');
  const [phone, setPhone] = useState('');
  const [order, setOrder] = useState<Order | null>(null);

  const checkout = useMutation({
    mutationFn: () => orderApi.checkout(address, phone),
    onSuccess: (o) => {
      setOrder(o);
      // Сервер очистил корзину и списал остатки товаров.
      qc.invalidateQueries({ queryKey: cartKeys.cart });
      qc.invalidateQueries({ queryKey: productKeys.all });
      qc.invalidateQueries({ queryKey: orderKeys.mine });
    },
  });

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose();
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  const items = cart?.items ?? [];
  const unavailable = items.some((i) => !i.available);

  return (
    <div
      className="fixed inset-0 z-[100] flex justify-end bg-slate-950/40 backdrop-blur-xs animate-in fade-in duration-200"
      onClick={onClose}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-label="Корзина покупок"
        onClick={(e) => e.stopPropagation()}
        className="w-full sm:w-[460px] bg-white h-full shadow-2xl flex flex-col justify-between p-6 overflow-y-auto"
      >
        <div className="space-y-6">
          <div className="flex items-center justify-between pb-4 border-b border-slate-200">
            <div className="flex items-center gap-2">
              <ShoppingBag className="w-5 h-5 text-slate-900" />
              <h2 className="text-lg font-bold text-slate-950">Корзина покупок</h2>
            </div>
            <button type="button" onClick={onClose} aria-label="Закрыть корзину" className="p-1 rounded-lg text-slate-400 hover:text-slate-900">
              <X className="w-5 h-5" />
            </button>
          </div>

          {error && <ErrorMessage error={error} />}

          {order ? (
            <div className="p-6 text-center space-y-3">
              <CheckCircle2 className="w-12 h-12 text-emerald-600 mx-auto" />
              <h3 className="text-lg font-bold text-slate-950">Заказ № {order.id} оформлен</h3>
              <p className="text-xs text-slate-600">Сумма: {rub(order.totalKopecks)}.</p>
              {order.paymentUrl ? (
                <a
                  href={order.paymentUrl}
                  className="block w-full py-4 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-black uppercase tracking-wider"
                >
                  Перейти к оплате
                </a>
              ) : (
                <p className="text-xs text-slate-500">
                  Онлайн-оплата сейчас не подключена — заказ ждёт подтверждения администратором.
                </p>
              )}
              <Link href="/account?tab=orders" className="block text-xs font-bold text-blue-600 hover:text-blue-800">
                Мои заказы →
              </Link>
            </div>
          ) : isLoading ? (
            <Loading />
          ) : items.length === 0 ? (
            <div className="p-12 text-center text-slate-400 text-sm">Ваша корзина пока пуста.</div>
          ) : (
            <ul className="space-y-3">
              {items.map((item) => (
                <li key={item.productId} className="flex items-center justify-between gap-3 p-3 bg-slate-50 rounded-xl border border-slate-200">
                  <div className="min-w-0 flex-1">
                    <h3 className="text-xs font-bold text-slate-950 truncate">{item.title}</h3>
                    <div className="text-[11px] text-slate-500">{rub(item.priceKopecks)} за шт.</div>
                    {!item.available && <div className="text-[11px] font-semibold text-rose-600">Осталось только {item.stock} шт.</div>}
                  </div>
                  <div className="flex items-center gap-1.5">
                    <button
                      type="button"
                      disabled={isUpdating || item.quantity <= 1}
                      onClick={() => setQuantity(item.productId, item.quantity - 1).catch(() => undefined)}
                      aria-label="Меньше"
                      className="p-1 rounded border border-slate-200 disabled:opacity-40"
                    >
                      <Minus className="w-3 h-3" />
                    </button>
                    <span className="w-6 text-center text-xs font-bold">{item.quantity}</span>
                    <button
                      type="button"
                      disabled={isUpdating || item.quantity >= item.stock}
                      onClick={() => setQuantity(item.productId, item.quantity + 1).catch(() => undefined)}
                      aria-label="Больше"
                      className="p-1 rounded border border-slate-200 disabled:opacity-40"
                    >
                      <Plus className="w-3 h-3" />
                    </button>
                  </div>
                  <span className="text-xs font-bold text-slate-900 font-mono w-20 text-right">{rub(item.lineTotalKopecks)}</span>
                  <button
                    type="button"
                    disabled={isUpdating}
                    onClick={() => remove(item.productId).catch(() => undefined)}
                    aria-label={`Убрать «${item.title}» из корзины`}
                    className="text-slate-400 hover:text-rose-600 transition-colors"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>

        {!order && items.length > 0 && cart && (
          <form
            className="pt-6 border-t border-slate-200 space-y-4"
            onSubmit={(e) => {
              e.preventDefault();
              checkout.mutate();
            }}
          >
            <Field label="Адрес доставки или «самовывоз»" required>
              <Input required maxLength={500} value={address} onChange={(e) => setAddress(e.target.value)} />
            </Field>
            <Field label="Телефон" required>
              <Input
                type="tel"
                required
                placeholder="+7 900 000-00-00"
                // Как на сервере: цифры, пробелы, скобки, дефисы, «+» в начале.
                pattern="\+?[0-9 ()\-]{10,20}"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
              />
            </Field>
            {checkout.error && <ErrorMessage error={checkout.error} />}
            <div className="flex items-center justify-between text-base font-bold text-slate-950">
              <span>Итого к оплате:</span>
              <span className="font-mono text-xl">{rub(cart.totalKopecks)}</span>
            </div>
            <Button type="submit" disabled={checkout.isPending || unavailable}>
              {checkout.isPending ? 'Оформляем...' : unavailable ? 'Уменьшите количество' : 'Оформить заказ'}
            </Button>
          </form>
        )}
      </div>
    </div>
  );
}
