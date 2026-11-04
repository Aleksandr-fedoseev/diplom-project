export interface ProductItem {
  id: string;
  title: string;
  category: 'clothes' | 'accessories' | 'hats' | 'books' | 'prints';
  categoryLabel: string;
  price: number;
  coverImage: string;
  description: string;
  isPopular?: boolean;
  inStock: boolean;
}
