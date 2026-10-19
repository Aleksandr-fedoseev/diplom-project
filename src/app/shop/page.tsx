'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { 
  ShoppingBag, 
  Heart, 
  ShieldCheck, 
  BookOpen, 
  Check, 
  ArrowRight,
  Sparkles,
  CheckCircle2,
  Trash2,
  X
} from 'lucide-react';
import { PRODUCTS } from '@/data/shop';
import { ProductItem } from '@/types/content';

export default function ShopPage() {
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [cart, setCart] = useState<{ product: ProductItem; quantity: number }[]>([]);
  const [isCartOpen, setIsCartOpen] = useState(false);
  const [orderSubmitted, setOrderSubmitted] = useState(false);

  const categories = [
    { id: 'all', label: 'Все товары' },
    { id: 'clothes', label: 'Одежда' },
    { id: 'accessories', label: 'Аксессуары' },
    { id: 'hats', label: 'Головные уборы' },
    { id: 'books', label: 'Книги и издания' },
  ];

  const filtered = PRODUCTS.filter((item) => 
    selectedCategory === 'all' || item.category === selectedCategory
  );

  const addToCart = (product: ProductItem) => {
    setCart((prev) => {
      const existing = prev.find((item) => item.product.id === product.id);
      if (existing) {
        return prev.map((item) =>
          item.product.id === product.id ? { ...item, quantity: item.quantity + 1 } : item
        );
      }
      return [...prev, { product, quantity: 1 }];
    });
    setIsCartOpen(true);
  };

  const removeFromCart = (id: string) => {
    setCart((prev) => prev.filter((item) => item.product.id !== id));
  };

  const totalAmount = cart.reduce((sum, item) => sum + item.product.price * item.quantity, 0);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      
      {/* Header & Hero */}
      <div className="bg-slate-950 text-white rounded-3xl p-8 sm:p-12 mb-12 shadow-md">
        <div className="max-w-2xl space-y-3">
          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400 bg-white/10 px-3 py-1 rounded">
            Фирменная сувенирная продукция
          </span>
          <h1 className="text-3xl sm:text-5xl font-black tracking-tight leading-tight text-white">
            Сувенирная продукция
          </h1>
          <p className="text-base sm:text-lg text-slate-200 font-normal leading-relaxed">
            Носи историю. Поддержи проект. Все средства от реализации продукции направляются на развитие интерактивной платформы, архивные исследования и цифровую фиксацию культурного наследия региона.
          </p>

          <div className="pt-2 flex items-center gap-3">
            <button
              onClick={() => setIsCartOpen(true)}
              className="inline-flex items-center gap-2 px-6 py-3.5 bg-white text-slate-950 rounded-xl text-xs font-bold uppercase tracking-wider hover:bg-slate-100 transition-colors shadow-sm"
            >
              <ShoppingBag className="w-4 h-4" />
              <span>Корзина ({cart.reduce((s, i) => s + i.quantity, 0)})</span>
            </button>
          </div>
        </div>
      </div>

      {/* Category Pills */}
      <div className="flex items-center gap-2 overflow-x-auto pb-4 mb-8 border-b border-slate-200 scrollbar-none">
        {categories.map((cat) => (
          <button
            key={cat.id}
            onClick={() => setSelectedCategory(cat.id)}
            className={`px-4 py-2 rounded-xl text-xs font-bold uppercase tracking-wider whitespace-nowrap transition-colors ${
              selectedCategory === cat.id
                ? 'bg-slate-950 text-white shadow-xs'
                : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
            }`}
          >
            {cat.label}
          </button>
        ))}
      </div>

      {/* Product Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 mb-16">
        {filtered.map((product) => (
          <div
            key={product.id}
            className="bg-white border border-slate-200 rounded-2xl p-6 shadow-xs hover:border-slate-400 transition-all flex flex-col justify-between group"
          >
            <div className="space-y-3">
              <span className="inline-block px-2.5 py-1 bg-slate-100 text-slate-700 text-[10px] font-bold uppercase tracking-wider rounded">
                {product.categoryLabel}
              </span>
              <h3 className="text-base sm:text-lg font-bold text-slate-950 leading-snug">
                {product.title}
              </h3>
              <div className="text-xl font-black text-slate-950 pt-1">
                {product.price.toLocaleString('ru-RU')} ₽
              </div>
            </div>

            <div className="pt-6">
              <button
                onClick={() => addToCart(product)}
                className="w-full py-3 px-4 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-bold uppercase tracking-wider flex items-center justify-center gap-2 transition-colors shadow-xs"
              >
                <ShoppingBag className="w-4 h-4" />
                <span>Добавить в корзину</span>
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Trust Block: «Больше, чем сувениры» (ТЗ 9 и Скриншот 3) */}
      <div className="bg-slate-50 border border-slate-200 rounded-3xl p-8 sm:p-12 mb-16 space-y-8">
        <div className="max-w-xl">
          <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Миссия фонда</span>
          <h2 className="text-2xl sm:text-3xl font-black text-slate-950 mt-1">
            Больше, чем сувениры
          </h2>
          <p className="text-xs sm:text-sm text-slate-600 mt-2 leading-relaxed">
            Покупая сувенирную продукцию, вы напрямую инвестируете в сохранение исторической памяти родного края.
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs space-y-2">
            <div className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center font-bold text-slate-900 text-sm">
              01
            </div>
            <h4 className="text-sm font-bold text-slate-900">Развиваем платформу</h4>
            <p className="text-xs text-slate-500 leading-relaxed">
              Средства идут на фотофиксацию, создание 3D-панорам и оцифровку архитектурных памятников.
            </p>
          </div>

          <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs space-y-2">
            <div className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center font-bold text-slate-900 text-sm">
              02
            </div>
            <h4 className="text-sm font-bold text-slate-900">Поддерживаем мастеров</h4>
            <p className="text-xs text-slate-500 leading-relaxed">
              Все изделия шьются и отпечатываются на локальных астраханских производствах.
            </p>
          </div>

          <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs space-y-2">
            <div className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center font-bold text-slate-900 text-sm">
              03
            </div>
            <h4 className="text-sm font-bold text-slate-900">Архивные исследования</h4>
            <p className="text-xs text-slate-500 leading-relaxed">
              Финансируем работу историков и архивистов по изучению первоисточников в фондах ГАОО.
            </p>
          </div>

          <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs space-y-2">
            <div className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center font-bold text-slate-900 text-sm">
              04
            </div>
            <h4 className="text-sm font-bold text-slate-900">Делаем край ближе</h4>
            <p className="text-xs text-slate-500 leading-relaxed">
              Красивый способ рассказать о подлинной культуре Нижней Волги друзьям и гостям города.
            </p>
          </div>
        </div>
      </div>

      {/* Cart Modal / Slide-over Drawer */}
      {isCartOpen && (
        <div className="fixed inset-0 z-[100] flex justify-end bg-slate-950/40 backdrop-blur-xs animate-in fade-in duration-200">
          <div className="w-full sm:w-[460px] bg-white h-full shadow-2xl flex flex-col justify-between p-6 overflow-y-auto">
            
            <div className="space-y-6">
              <div className="flex items-center justify-between pb-4 border-b border-slate-200">
                <div className="flex items-center gap-2">
                  <ShoppingBag className="w-5 h-5 text-slate-900" />
                  <h3 className="text-lg font-bold text-slate-950">Корзина покупок</h3>
                </div>
                <button
                  onClick={() => setIsCartOpen(false)}
                  className="p-1 rounded-lg text-slate-400 hover:text-slate-900"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              {orderSubmitted ? (
                <div className="p-8 text-center space-y-3">
                  <CheckCircle2 className="w-12 h-12 text-emerald-600 mx-auto" />
                  <h4 className="text-lg font-bold text-slate-950">Заказ успешно оформлен!</h4>
                  <p className="text-xs text-slate-600">
                    Спасибо за поддержку фонда! Менеджер свяжется с вами для согласования самовывоза или доставки.
                  </p>
                  <button
                    onClick={() => {
                      setCart([]);
                      setOrderSubmitted(false);
                      setIsCartOpen(false);
                    }}
                    className="w-full py-3 bg-slate-950 text-white rounded-xl text-xs font-bold uppercase tracking-wider"
                  >
                    Вернуться в магазин
                  </button>
                </div>
              ) : cart.length === 0 ? (
                <div className="p-12 text-center text-slate-400 text-sm">
                  Ваша корзина пока пуста.
                </div>
              ) : (
                <div className="space-y-4">
                  {cart.map((item) => (
                    <div key={item.product.id} className="flex items-center justify-between gap-3 p-3 bg-slate-50 rounded-xl border border-slate-200">
                      <div className="min-w-0 flex-1">
                        <h4 className="text-xs font-bold text-slate-950 truncate">{item.product.title}</h4>
                        <div className="text-[11px] text-slate-500">{item.product.price} ₽ × {item.quantity} шт.</div>
                      </div>
                      <div className="flex items-center gap-3">
                        <span className="text-xs font-bold text-slate-900 font-mono">
                          {item.product.price * item.quantity} ₽
                        </span>
                        <button
                          onClick={() => removeFromCart(item.product.id)}
                          className="text-slate-400 hover:text-rose-600 transition-colors"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {!orderSubmitted && cart.length > 0 && (
              <div className="pt-6 border-t border-slate-200 space-y-4">
                <div className="flex items-center justify-between text-base font-bold text-slate-950">
                  <span>Итого к оплате:</span>
                  <span className="font-mono text-xl">{totalAmount.toLocaleString('ru-RU')} ₽</span>
                </div>
                <button
                  onClick={() => setOrderSubmitted(true)}
                  className="w-full py-3.5 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-black uppercase tracking-wider transition-colors"
                >
                  Оформить заказ
                </button>
              </div>
            )}

          </div>
        </div>
      )}

    </div>
  );
}
