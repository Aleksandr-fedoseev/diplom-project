package shop

import (
	"context"
	"errors"
)

// Порты сервиса: всё, что ему нужно от техники и других модулей. Типы — свои,
// чтобы сервис не зависел ни от клиента ЮKassa, ни от модуля auth. Реализации
// подключаются при сборке приложения (адаптеры в app), в тестах — фейки.

// TxManager выполняет fn в одной транзакции (реализация — infrastructure/postgres.TxManager).
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

// EventPublisher записывает событие для других модулей (outbox) в текущую транзакцию.
type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload any) error // payload кодируется в JSON
}

// Users — то, что магазину нужно от модуля пользователей. auth.Service подходит без адаптера.
type Users interface {
	// EnsureActive возвращает ошибку домена (403), если пользователь заблокирован.
	EnsureActive(ctx context.Context, id int64) error
}

// PaymentGateway — платёжный сервис (ЮKassa).
type PaymentGateway interface {
	// CreatePayment создаёт платёж и возвращает его id и страницу оплаты. idempotenceKey
	// защищает от второго платежа, если ответ потерялся и запрос повторили.
	CreatePayment(ctx context.Context, orderID, amountKopecks int64, description, returnURL, idempotenceKey string) (paymentID, confirmationURL string, err error)
	// GetPayment запрашивает платёж по id — его настоящий статус и сумму.
	GetPayment(ctx context.Context, id string) (Payment, error)
}

// Payment — платёж в том виде, который нужен магазину.
type Payment struct {
	ID      string
	Status  string // PaymentSucceeded, PaymentCanceled или другой (ещё не завершён)
	Amount  string // сумма строкой, как её пишет ЮKassa: "4500.00"
	OrderID string // metadata.order_id — только для логов
}

// Статусы платежа, на которые реагирует магазин.
const (
	PaymentSucceeded = "succeeded"
	PaymentCanceled  = "canceled"
)

// Ошибки платёжного шлюза. Адаптер переводит в них ошибки клиента ЮKassa.
var (
	// ErrGatewayDisabled — ЮKassa не настроена: заказы оформляются без оплаты.
	ErrGatewayDisabled = errors.New("приём оплаты не настроен")
	// ErrGatewayPaymentNotFound — ЮKassa не знает платежа с таким id.
	ErrGatewayPaymentNotFound = errors.New("платёж не найден")
)

// События магазина (контракт с модулем уведомлений).
const (
	EventOrderPlaced    = "order.placed"
	EventOrderPaid      = "order.paid"
	EventOrderShipped   = "order.shipped"
	EventOrderCancelled = "order.cancelled"
)

// OrderPlaced — payload order.placed. Адрес и позиции нужны, чтобы письмо перечислило покупки.
type OrderPlaced struct {
	OrderID         int64             `json:"orderId"`
	UserID          int64             `json:"userId"`
	TotalKopecks    int64             `json:"totalKopecks"`
	DeliveryAddress string            `json:"deliveryAddress"`
	Items           []OrderPlacedItem `json:"items"`
}

// OrderPlacedItem — позиция заказа в событии order.placed (снимок названия и цены).
type OrderPlacedItem struct {
	Title        string `json:"title"`
	Quantity     int    `json:"quantity"`
	PriceKopecks int64  `json:"priceKopecks"`
}

// OrderPaid — payload order.paid.
type OrderPaid struct {
	OrderID      int64 `json:"orderId"`
	UserID       int64 `json:"userId"`
	TotalKopecks int64 `json:"totalKopecks"`
}

// OrderShipped — payload order.shipped.
type OrderShipped struct {
	OrderID         int64  `json:"orderId"`
	UserID          int64  `json:"userId"`
	DeliveryAddress string `json:"deliveryAddress"`
}

// OrderCancelled — payload order.cancelled. WasPaid — заказ был оплачен (письмо обещает возврат денег).
type OrderCancelled struct {
	OrderID      int64 `json:"orderId"`
	UserID       int64 `json:"userId"`
	TotalKopecks int64 `json:"totalKopecks"`
	WasPaid      bool  `json:"wasPaid"`
}
