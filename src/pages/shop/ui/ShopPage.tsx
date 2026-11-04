'use client';

import { useCallback, useState } from 'react';
import { ShoppingBag } from 'lucide-react';
import { PRODUCT_CATEGORIES, PRODUCTS, ProductCard } from '@/entities/product';
import { Chip } from '@/shared/ui';
import { addToCart, cartCount, removeFromCart, type CartLine } from '../model/cart';
import { CartDrawer } from './CartDrawer';
import { MissionBlock } from './MissionBlock';

const CATEGORIES = [{ id: 'all', label: 'Все товары' }, ...PRODUCT_CATEGORIES];

export default function ShopPage() {
  const [category, setCategory] = useState('all');
  const [cart, setCart] = useState<CartLine[]>([]);
  const [isCartOpen, setIsCartOpen] = useState(false);
  const [ordered, setOrdered] = useState(false);
  const closeCart = useCallback(() => setIsCartOpen(false), []);

  // Рубрики без товаров не показываем.
  const categories = CATEGORIES.filter((c) => c.id === 'all' || PRODUCTS.some((p) => p.category === c.id));
  const products = PRODUCTS.filter((p) => category === 'all' || p.category === category);

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
          <div className="pt-2">
            <button
              type="button"
              onClick={() => setIsCartOpen(true)}
              className="inline-flex items-center gap-2 px-6 py-3.5 bg-white text-slate-950 rounded-xl text-xs font-bold uppercase tracking-wider hover:bg-slate-100 transition-colors shadow-sm"
            >
              <ShoppingBag className="w-4 h-4" />
              <span>Корзина ({cartCount(cart)})</span>
            </button>
          </div>
        </div>
      </div>

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

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 mb-16">
        {products.map((product) => (
          <ProductCard
            key={product.id}
            product={product}
            action={
              <button
                type="button"
                onClick={() => {
                  setCart((c) => addToCart(c, product));
                  setIsCartOpen(true);
                }}
                className="w-full py-3 px-4 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-bold uppercase tracking-wider flex items-center justify-center gap-2 transition-colors shadow-xs"
              >
                <ShoppingBag className="w-4 h-4" />
                <span>Добавить в корзину</span>
              </button>
            }
          />
        ))}
      </div>

      <MissionBlock />

      {isCartOpen && (
        <CartDrawer
          cart={cart}
          ordered={ordered}
          onRemove={(id) => setCart((c) => removeFromCart(c, id))}
          onCheckout={() => setOrdered(true)}
          onClose={closeCart}
          onFinish={() => {
            setCart([]);
            setOrdered(false);
            setIsCartOpen(false);
          }}
        />
      )}
    </div>
  );
}
