import type { ProductItem } from './types';

export const PRODUCT_CATEGORIES: { id: ProductItem['category']; label: string }[] = [
  { id: 'clothes', label: 'Одежда' },
  { id: 'accessories', label: 'Аксессуары' },
  { id: 'hats', label: 'Головные уборы' },
  { id: 'books', label: 'Книги и издания' },
  { id: 'prints', label: 'Полиграфия' },
];
