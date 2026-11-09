import { api, mediaUrl } from '@/shared/api';
import { productCategoryLabel } from '../model/categories';
import type { ProductCategory, ProductDto, ProductInput, ProductItem } from '../model/types';

export const fromProductDto = (p: ProductDto): ProductItem => ({
  id: p.id,
  slug: p.slug,
  title: p.title,
  category: p.category as ProductCategory,
  categoryLabel: productCategoryLabel(p.category),
  price: p.priceKopecks / 100,
  // Стартовые товары ссылаются на картинки сайта (/merch/...), загруженные — на API.
  coverImage: p.imageUrl.startsWith('/api/') ? mediaUrl(p.imageUrl) : p.imageUrl,
  description: p.description,
  stock: p.stock,
  inStock: p.inStock,
  isActive: p.isActive,
});

export const productKeys = { all: ['products'] as const, list: (all: boolean) => ['products', { all }] as const };

export const productApi = {
  // all=true — вместе со снятыми с продажи (сервер выполнит только для администратора).
  list: async (all = false) =>
    (await api<ProductDto[]>('/products', { query: { all: all || undefined } })).map(fromProductDto),
  create: (input: ProductInput) => api<ProductDto>('/products', { method: 'POST', json: input }),
  update: (id: number, input: ProductInput) => api<ProductDto>(`/products/${id}`, { method: 'PUT', json: input }),
  remove: (id: number) => api<void>(`/products/${id}`, { method: 'DELETE' }),
};
