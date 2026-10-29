// Package notification — email-уведомления пользователей о событиях других модулей.
// Модуль не хранит данных: он получает события из outbox (заявка одобрена, заказ оплачен…),
// находит получателя и отправляет письмо.
package notification

// Имена событий, на которые подписан модуль. Модули-производители объявляют
// те же строки у себя (импортировать чужой domain нельзя) — это контракт между модулями.
const (
	EventSubmissionApproved = "submission.approved"
	EventSubmissionRejected = "submission.rejected"
	EventOrderPlaced        = "order.placed"
	EventOrderPaid          = "order.paid"
	EventOrderShipped       = "order.shipped"
	EventOrderCancelled     = "order.cancelled"
)

// Данные событий (payload в JSON). Теги json — часть контракта, менять их можно
// только вместе с модулями-производителями.

// SubmissionApproved — модератор одобрил заявку.
type SubmissionApproved struct {
	SubmissionID int64  `json:"submissionId"`
	AuthorID     int64  `json:"authorId"`
	Title        string `json:"title"`
	ObjectSlug   string `json:"objectSlug"` // объект, созданный или дополненный заявкой
}

// SubmissionRejected — модератор отклонил заявку (и, возможно, заблокировал автора).
type SubmissionRejected struct {
	SubmissionID  int64  `json:"submissionId"`
	AuthorID      int64  `json:"authorId"`
	Title         string `json:"title"`
	Reason        string `json:"reason"`
	AuthorBlocked bool   `json:"authorBlocked"`
}

// OrderPlaced — покупатель оформил заказ. UserID = 0 — покупатель удалён.
// DeliveryAddress и Items необязательны: без них письмо просто короче.
type OrderPlaced struct {
	OrderID         int64       `json:"orderId"`
	UserID          int64       `json:"userId"`
	TotalKopecks    int64       `json:"totalKopecks"`
	DeliveryAddress string      `json:"deliveryAddress,omitempty"`
	Items           []OrderItem `json:"items,omitempty"`
}

// OrderItem — строка заказа: товар, количество и цена за штуку на момент заказа.
type OrderItem struct {
	Title        string `json:"title"`
	Quantity     int64  `json:"quantity"`
	PriceKopecks int64  `json:"priceKopecks"`
}

// OrderPaid — оплата заказа получена.
type OrderPaid struct {
	OrderID      int64 `json:"orderId"`
	UserID       int64 `json:"userId"`
	TotalKopecks int64 `json:"totalKopecks"`
}

// OrderShipped — заказ передан в доставку.
type OrderShipped struct {
	OrderID         int64  `json:"orderId"`
	UserID          int64  `json:"userId"`
	DeliveryAddress string `json:"deliveryAddress"`
}

// OrderCancelled — заказ отменён; WasPaid — деньги нужно вернуть.
type OrderCancelled struct {
	OrderID      int64 `json:"orderId"`
	UserID       int64 `json:"userId"`
	TotalKopecks int64 `json:"totalKopecks"`
	WasPaid      bool  `json:"wasPaid"`
}
