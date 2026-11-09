export type OrderStatus = 'pending' | 'paid' | 'shipped' | 'completed' | 'cancelled';

export interface OrderItem {
  productId: number;
  title: string;
  quantity: number;
  priceKopecks: number;
  lineTotalKopecks: number;
}

export interface Order {
  id: number;
  userId?: number;
  status: OrderStatus;
  totalKopecks: number;
  paymentUrl: string; // страница оплаты ЮKassa; пусто — платёж не создан
  deliveryAddress: string;
  phone: string;
  items: OrderItem[];
  createdAt: string;
  updatedAt: string;
}
