'use client';

import { useCallback, useState } from 'react';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { ShoppingBag } from 'lucide-react';
import { useCart } from '@/entities/cart';
import { PRODUCT_CATEGORIES, productApi, productKeys, ProductCard, type ProductItem } from '@/entities/product';
import { useSession } from '@/entities/session';
import { isStaff } from '@/entities/user';
import { Chip, ErrorMessage, QueryState } from '@/shared/ui';
import { CartDrawer } from './CartDrawer';
import { MissionBlock } from './MissionBlock';

const CATEGORIES = [{ id: 'all', label: 'Все товары' }, ...PRODUCT_CATEGORIES];

const buttonClass =
  'w-full py-3 px-4 bg-slate-950 hover:bg-slate-800 disabled:bg-slate-300 text-white rounded-xl text-xs font-bold uppercase tracking-wider flex items-center justify-center gap-2 transition-colors shadow-xs';

export default function ShopPage() {
  const [category, setCategory] = useState('all');
  const [isCartOpen, setIsCartOpen] = useState(false);
  const closeCart = useCallback(() => setIsCartOpen(false), []);
  const products = useQuery({ queryKey: productKeys.list(false), queryFn: () => productApi.list() });
  const { isAuthed, user } = useSession();
  const staff = isStaff(user?.role);
  const cart = useCart();

  const all = products.data ?? [];
  // Рубрики без товаров не показываем.
  const categories = CATEGORIES.filter((c) => c.id === 'all' || all.some((p) => p.category === c.id));

  const action = (product: ProductItem) => {
    if (!product.inStock) {
      return (
        <button type="button" disabled className={buttonClass}>
          Нет в наличии
        </button>
      );
    }
    // Сотрудник покупает с личного аккаунта: в рабочем нет раздела «Мои заказы».
    if (staff) {
      return (
        <button type="button" disabled className={buttonClass} title="Покупки — с обычного аккаунта">
          Покупки — с личного аккаунта
        </button>
      );
    }
    // Корзина хранится на сервере у пользователя, поэтому покупать можно только после входа.
    if (!isAuthed) {
      return (
        <Link href="/login?next=/shop" className={buttonClass}>
          Войдите, чтобы купить
        </Link>
      );
    }
    return (
      <button
        type="button"
        disabled={cart.isUpdating}
        onClick={async () => {
          await cart.add(product.id).catch(() => undefined); // ошибку показывает корзина
          setIsCartOpen(true);
        }}
        className={buttonClass}
      >
        <ShoppingBag className="w-4 h-4" />
        <span>Добавить в корзину</span>
      </button>
    );
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      <div className="bg-slate-950 text-white rounded-3xl p-8 sm:p-12 mb-12 shadow-md">
        <div className="max-w-2xl space-y-3">
          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400 bg-white/10 px-3 py-1 rounded">
            Фирменная сувенирная продукция
          </span>
          <h1 className="text-3xl sm:text-5xl font-black tracking-tight leading-tight text-white">Сувенирная продукция</h1>
          <p className="text-base sm:text-lg text-slate-200 font-normal leading-relaxed">
            Носи историю. Поддержи проект. Все средства от реализации продукции направляются на развитие интерактивной
            платформы, архивные исследования и цифровую фиксацию культурного наследия региона.
          </p>
          {isAuthed && !staff && (
            <div className="pt-2 flex flex-wrap gap-3">
              <button
                type="button"
                onClick={() => setIsCartOpen(true)}
                className="inline-flex items-center gap-2 px-6 py-3.5 bg-white text-slate-950 rounded-xl text-xs font-bold uppercase tracking-wider hover:bg-slate-100 transition-colors shadow-sm"
              >
                <ShoppingBag className="w-4 h-4" />
                <span>Корзина ({cart.count})</span>
              </button>
              <Link
                href="/account?tab=orders"
                className="inline-flex items-center px-6 py-3.5 border border-white/30 text-white rounded-xl text-xs font-bold uppercase tracking-wider hover:bg-white/10 transition-colors"
              >
                Мои заказы
              </Link>
            </div>
          )}
        </div>
      </div>

      <QueryState query={products} loadingLabel="Загрузка товаров...">
        {() => (
          <>
            <div className="flex items-center gap-2 overflow-x-auto pb-4 mb-8 border-b border-slate-200 scrollbar-none">
              {categories.map((c) => (
                <Chip
                  key={c.id}
                  active={category === c.id}
                  onClick={() => setCategory(c.id)}
                  className="px-4 py-2 text-xs uppercase tracking-wider whitespace-nowrap"
                >
                  {c.label}
                </Chip>
              ))}
            </div>
            {!isCartOpen && cart.error && <ErrorMessage className="mb-6" error={cart.error} />}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 mb-16">
              {all
                .filter((p) => category === 'all' || p.category === category)
                .map((product) => (
                  <ProductCard key={product.id} product={product} action={action(product)} />
                ))}
            </div>
          </>
        )}
      </QueryState>

      <MissionBlock />

      {isCartOpen && <CartDrawer onClose={closeCart} />}
    </div>
  );
}
