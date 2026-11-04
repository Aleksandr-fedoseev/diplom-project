import { describe, expect, it } from 'vitest';
import { PRODUCTS } from '@/entities/product';
import { addToCart, cartCount, cartTotal, removeFromCart } from './cart';

const [a, b] = PRODUCTS;

describe('корзина', () => {
  it('повторное добавление увеличивает количество, а не дублирует строку', () => {
    const cart = addToCart(addToCart([], a), a);
    expect(cart).toEqual([{ product: a, quantity: 2 }]);
  });

  it('считает сумму и число товаров', () => {
    const cart = addToCart(addToCart(addToCart([], a), a), b);
    expect(cartCount(cart)).toBe(3);
    expect(cartTotal(cart)).toBe(a.price * 2 + b.price);
  });

  it('удаляет строку целиком и не трогает остальные', () => {
    const cart = removeFromCart(addToCart(addToCart([], a), b), a.id);
    expect(cart).toEqual([{ product: b, quantity: 1 }]);
  });

  it('не меняет исходный массив', () => {
    const cart = addToCart([], a);
    addToCart(cart, a);
    expect(cart[0].quantity).toBe(1);
  });

  it('пустая корзина — ноль', () => {
    expect(cartTotal([])).toBe(0);
    expect(cartCount([])).toBe(0);
  });
});
