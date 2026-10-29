// Package shop — сценарии магазина: товары, корзина, оформление заказа, оплата через
// платёжный шлюз и смена статусов. Реализует shop.Service из домена.
// Писем модуль не отправляет: он публикует события (order.placed, order.paid …),
// а письма по ним пишет модуль уведомлений.
package shop

import (
	"context"

	"heritage-backend/internal/domain/shop"
)

// Deps — зависимости сервиса. Payments может быть nil: тогда, как и с ненастроенной
// ЮKassa, заказы оформляются без платежа. ReturnURL — страница сайта, куда ЮKassa
// вернёт покупателя; к ней добавляется ?order=<id>.
type Deps struct {
	Tx        TxManager
	Products  shop.ProductRepository
	Carts     shop.CartRepository
	Orders    shop.OrderRepository
	Users     Users
	Events    EventPublisher
	Payments  PaymentGateway
	ReturnURL string
}

type Service struct {
	tx        TxManager
	products  shop.ProductRepository
	carts     shop.CartRepository
	orders    shop.OrderRepository
	users     Users
	events    EventPublisher
	payments  PaymentGateway
	returnURL string
}

// Проверка при сборке: Service реализует интерфейс домена.
var _ shop.Service = (*Service)(nil)

func NewService(d Deps) *Service {
	return &Service{
		tx: d.Tx, products: d.Products, carts: d.Carts, orders: d.Orders,
		users: d.Users, events: d.Events, payments: d.Payments, returnURL: d.ReturnURL,
	}
}

// Товары.

func (s *Service) ListProducts(ctx context.Context, category string, includeInactive bool) ([]shop.Product, error) {
	return s.products.List(ctx, category, includeInactive)
}

func (s *Service) GetProduct(ctx context.Context, slug string, includeInactive bool) (shop.Product, error) {
	p, err := s.products.BySlug(ctx, slug)
	if err != nil {
		return shop.Product{}, err
	}
	if !p.IsActive && !includeInactive {
		return shop.Product{}, shop.ErrProductNotFound
	}
	return p, nil
}

func (s *Service) CreateProduct(ctx context.Context, in shop.ProductInput) (shop.Product, error) {
	if err := in.Normalize(); err != nil {
		return shop.Product{}, err
	}
	return s.products.Create(ctx, in.Product(0))
}

// UpdateProduct меняет товар. Новая цена действует для новых заказов:
// в оформленных заказах остаётся цена на момент покупки.
func (s *Service) UpdateProduct(ctx context.Context, id int64, in shop.ProductInput) (shop.Product, error) {
	if err := in.Normalize(); err != nil {
		return shop.Product{}, err
	}
	return s.products.Update(ctx, in.Product(id))
}

// DeactivateProduct снимает товар с продажи. Физически товар не удаляется:
// на него ссылаются позиции оформленных заказов. Из корзин товар убирается.
func (s *Service) DeactivateProduct(ctx context.Context, id int64) error {
	return s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.products.Deactivate(ctx, id); err != nil {
			return err
		}
		return s.carts.RemoveProduct(ctx, id)
	})
}

// Корзина.

func (s *Service) GetCart(ctx context.Context, userID int64) (shop.Cart, error) {
	items, err := s.carts.Items(ctx, userID)
	if err != nil {
		return shop.Cart{}, err
	}
	return shop.Cart{Items: items}, nil
}

func (s *Service) SetCartItem(ctx context.Context, userID, productID int64, quantity int) (shop.Cart, error) {
	if quantity < 0 {
		return shop.Cart{}, shop.ErrInvalidQuantity
	}
	if quantity == 0 {
		return s.RemoveCartItem(ctx, userID, productID)
	}
	p, err := s.products.ByID(ctx, productID)
	if err != nil {
		return shop.Cart{}, err
	}
	if err := shop.CheckCartQuantity(p, quantity); err != nil {
		return shop.Cart{}, err
	}
	if err := s.carts.Set(ctx, userID, productID, quantity); err != nil {
		return shop.Cart{}, err
	}
	return s.GetCart(ctx, userID)
}

func (s *Service) RemoveCartItem(ctx context.Context, userID, productID int64) (shop.Cart, error) {
	if err := s.carts.Remove(ctx, userID, productID); err != nil {
		return shop.Cart{}, err
	}
	return s.GetCart(ctx, userID)
}
