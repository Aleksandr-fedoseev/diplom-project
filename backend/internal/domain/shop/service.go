package shop

import "context"

// Service — сценарии магазина. Реализация — service/shop; transport зависит только от этого интерфейса.
type Service interface {
	// ListProducts — витрина, по желанию одной категории; includeInactive — и снятые (для администратора).
	ListProducts(ctx context.Context, category string, includeInactive bool) ([]Product, error)
	// GetProduct — карточка товара. Снятый с продажи товар виден только при includeInactive.
	GetProduct(ctx context.Context, slug string, includeInactive bool) (Product, error)
	CreateProduct(ctx context.Context, in ProductInput) (Product, error)
	UpdateProduct(ctx context.Context, id int64, in ProductInput) (Product, error)
	// DeactivateProduct снимает товар с продажи и убирает его из корзин.
	DeactivateProduct(ctx context.Context, id int64) error

	GetCart(ctx context.Context, userID int64) (Cart, error)
	// SetCartItem задаёт количество товара в корзине; 0 убирает товар.
	SetCartItem(ctx context.Context, userID, productID int64, quantity int) (Cart, error)
	RemoveCartItem(ctx context.Context, userID, productID int64) (Cart, error)

	// Checkout оформляет заказ из корзины и пытается создать платёж (paymentUrl).
	Checkout(ctx context.Context, userID int64, in CheckoutInput) (Order, error)
	// Pay возвращает заказ со ссылкой на оплату, создавая платёж, если его ещё нет.
	Pay(ctx context.Context, orderID, userID int64) (Order, error)
	// CancelMine — покупатель отменяет свой неоплаченный заказ.
	CancelMine(ctx context.Context, orderID, userID int64) (Order, error)
	GetOrder(ctx context.Context, id int64, viewer Viewer) (Order, error)
	ListMyOrders(ctx context.Context, userID int64) ([]Order, error)
	// ListOrders — все заказы для администратора; status "" — любой.
	ListOrders(ctx context.Context, status string, limit, offset int) ([]Order, error)
	// SetStatus — смена статуса администратором (только разрешённые переходы).
	SetStatus(ctx context.Context, id int64, status string) (Order, error)
	// HandlePaymentNotification обрабатывает уведомление ЮKassa о платеже paymentID.
	// Ошибка — только при временном сбое (тогда ЮKassa повторит уведомление).
	HandlePaymentNotification(ctx context.Context, paymentID string) error
}
