'use client';

import { useEffect } from 'react';
import { CheckCircle2, ShoppingBag, Trash2, X } from 'lucide-react';
import { formatRub } from '@/shared/lib';
import { Button } from '@/shared/ui';
import { cartTotal, type CartLine } from '../model/cart';

interface CartDrawerProps {
  cart: CartLine[];
  ordered: boolean;
  onRemove: (productId: string) => void;
  onCheckout: () => void;
  onClose: () => void;
  onFinish: () => void; // после оформления: очистить корзину и вернуться в магазин
}

// Корзина, выезжающая справа. Закрывается крестиком, кликом по фону и клавишей Esc.
export function CartDrawer({ cart, ordered, onRemove, onCheckout, onClose, onFinish }: CartDrawerProps) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose();
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

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

          {ordered ? (
            <div className="p-8 text-center space-y-3">
              <CheckCircle2 className="w-12 h-12 text-emerald-600 mx-auto" />
              <h3 className="text-lg font-bold text-slate-950">Заказ успешно оформлен!</h3>
              <p className="text-xs text-slate-600">
                Спасибо за поддержку фонда! Менеджер свяжется с вами для согласования самовывоза или доставки.
              </p>
              <Button variant="secondary" className="w-full" onClick={onFinish}>
                Вернуться в магазин
              </Button>
            </div>
          ) : cart.length === 0 ? (
            <div className="p-12 text-center text-slate-400 text-sm">Ваша корзина пока пуста.</div>
          ) : (
            <ul className="space-y-4">
              {cart.map(({ product, quantity }) => (
                <li key={product.id} className="flex items-center justify-between gap-3 p-3 bg-slate-50 rounded-xl border border-slate-200">
                  <div className="min-w-0 flex-1">
                    <h3 className="text-xs font-bold text-slate-950 truncate">{product.title}</h3>
                    <div className="text-[11px] text-slate-500">
                      {formatRub(product.price)} × {quantity} шт.
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-xs font-bold text-slate-900 font-mono">{formatRub(product.price * quantity)}</span>
                    <button
                      type="button"
                      onClick={() => onRemove(product.id)}
                      aria-label={`Убрать «${product.title}» из корзины`}
                      className="text-slate-400 hover:text-rose-600 transition-colors"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>

        {!ordered && cart.length > 0 && (
          <div className="pt-6 border-t border-slate-200 space-y-4">
            <div className="flex items-center justify-between text-base font-bold text-slate-950">
              <span>Итого к оплате:</span>
              <span className="font-mono text-xl">{formatRub(cartTotal(cart))}</span>
            </div>
            <Button onClick={onCheckout}>Оформить заказ</Button>
          </div>
        )}
      </div>
    </div>
  );
}
