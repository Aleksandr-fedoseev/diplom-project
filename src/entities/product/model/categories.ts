import type { ProductCategory } from './types';

export const PRODUCT_CATEGORIES: { id: ProductCategory; label: string }[] = [
  { id: 'clothes', label: 'Одежда' },
  { id: 'accessories', label: 'Аксессуары' },
  { id: 'hats', label: 'Головные уборы' },
  { id: 'books', label: 'Книги и издания' },
  { id: 'prints', label: 'Полиграфия' },
];

export const productCategoryLabel = (id: string) => PRODUCT_CATEGORIES.find((c) => c.id === id)?.label ?? id;
