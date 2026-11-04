import type { ProductItem } from '@/entities/product';

export interface CartLine {
  product: ProductItem;
  quantity: number;
}

// Корзина пока живёт на странице. На сервере она уже есть (/api/cart) — сюда придёт вместе с API.
export function addToCart(cart: CartLine[], product: ProductItem): CartLine[] {
  return cart.some((l) => l.product.id === product.id)
    ? cart.map((l) => (l.product.id === product.id ? { ...l, quantity: l.quantity + 1 } : l))
    : [...cart, { product, quantity: 1 }];
}

export function removeFromCart(cart: CartLine[], productId: string): CartLine[] {
  return cart.filter((l) => l.product.id !== productId);
}

export const cartTotal = (cart: CartLine[]) => cart.reduce((sum, l) => sum + l.product.price * l.quantity, 0);
export const cartCount = (cart: CartLine[]) => cart.reduce((sum, l) => sum + l.quantity, 0);
