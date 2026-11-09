// Корзина хранится на сервере у пользователя (internal/transport/http/shop).
export interface CartItem {
  productId: number;
  slug: string;
  title: string;
  imageUrl: string;
  priceKopecks: number;
  quantity: number;
  stock: number;
  lineTotalKopecks: number;
  available: boolean; // товар в продаже и его хватает
}

export interface Cart {
  items: CartItem[];
  totalKopecks: number;
}
