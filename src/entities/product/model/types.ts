export type ProductCategory = 'clothes' | 'accessories' | 'hats' | 'books' | 'prints';

export interface ProductItem {
  id: number;
  slug: string;
  title: string;
  category: ProductCategory;
  categoryLabel: string;
  price: number; // рубли; в API — копейки
  coverImage: string;
  description: string;
  stock: number;
  inStock: boolean;
  isActive: boolean;
}

// Формат API (internal/transport/http/shop).
export interface ProductDto {
  id: number;
  slug: string;
  title: string;
  description: string;
  category: string;
  priceKopecks: number;
  stock: number;
  inStock: boolean;
  imageUrl: string;
  isActive: boolean;
  updatedAt: string;
}

export interface ProductInput {
  slug: string;
  title: string;
  description: string;
  category: string;
  priceKopecks: number;
  stock: number;
  imageUrl: string;
  isActive?: boolean;
}
