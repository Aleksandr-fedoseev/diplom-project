import type { ReactNode } from 'react';
import { formatRub } from '@/shared/lib';
import type { ProductItem } from '../model/types';

// Карточка товара; action — кнопка действия (например, «В корзину»), её даёт страница.
export function ProductCard({ product, action }: { product: ProductItem; action: ReactNode }) {
  return (
    <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-xs hover:border-slate-400 transition-all flex flex-col justify-between group">
      <div className="space-y-3">
        <span className="inline-block px-2.5 py-1 bg-slate-100 text-slate-700 text-[10px] font-bold uppercase tracking-wider rounded">
          {product.categoryLabel}
        </span>
        <h3 className="text-base sm:text-lg font-bold text-slate-950 leading-snug">{product.title}</h3>
        <div className="text-xl font-black text-slate-950 pt-1">{formatRub(product.price)}</div>
      </div>
      <div className="pt-6">{action}</div>
    </div>
  );
}
