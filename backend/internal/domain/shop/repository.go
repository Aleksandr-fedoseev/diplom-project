package shop

import "context"

// Хранилища магазина. Реализация — repository/postgres. Методы берут транзакцию из ctx,
// если сервис открыл её через TxManager, поэтому несколько вызовов складываются в одну транзакцию.

// ProductRepository — товары и их остатки.
type ProductRepository interface {
	// List — товары по желанию одной категории ("" — все); includeInactive — и снятые с продажи.
	List(ctx context.Context, category string, includeInactive bool) ([]Product, error)
	BySlug(ctx context.Context, slug string) (Product, error) // ErrProductNotFound
	ByID(ctx context.Context, id int64) (Product, error)      // ErrProductNotFound
	Create(ctx context.Context, p Product) (Product, error)   // ErrSlugTaken
	Update(ctx context.Context, p Product) (Product, error)   // ErrProductNotFound, ErrSlugTaken
	// Deactivate снимает товар с продажи (физически не удаляет). ErrProductNotFound.
	Deactivate(ctx context.Context, id int64) error
	// DecrementStock уменьшает остаток; вызывается в транзакции оформления после LockForCheckout.
	DecrementStock(ctx context.Context, productID int64, quantity int) error
	// ReturnOrderStock возвращает на склад все позиции заказа (при отмене).
	ReturnOrderStock(ctx context.Context, orderID int64) error
}

// CartRepository — корзины покупателей.
type CartRepository interface {
	// Items — позиции корзины с текущими данными товаров, в порядке добавления.
	Items(ctx context.Context, userID int64) ([]CartItem, error)
	// Set задаёт количество (больше нуля), добавляя товар, если его ещё нет.
	Set(ctx context.Context, userID, productID int64, quantity int) error
	Remove(ctx context.Context, userID, productID int64) error
	Clear(ctx context.Context, userID int64) error
	// RemoveProduct убирает товар из всех корзин (товар сняли с продажи).
	RemoveProduct(ctx context.Context, productID int64) error
	// LockForCheckout возвращает позиции корзины и блокирует строки товаров (FOR UPDATE)
	// до конца транзакции — в порядке id товара, чтобы параллельные оформления не
	// попали во взаимную блокировку. Вызывать только внутри TxManager.Do.
	LockForCheckout(ctx context.Context, userID int64) ([]CheckoutLine, error)
}

// OrderRepository — заказы с позициями.
type OrderRepository interface {
	// Create записывает заказ и его позиции, возвращает id. Вызывать внутри TxManager.Do.
	Create(ctx context.Context, o Order) (int64, error)
	ByID(ctx context.Context, id int64) (Order, error)                // с позициями; ErrOrderNotFound
	ByPaymentID(ctx context.Context, paymentID string) (Order, error) // ErrOrderNotFound
	ListByUser(ctx context.Context, userID int64) ([]Order, error)    // новые сверху
	// List — все заказы для администратора; status "" — любой.
	List(ctx context.Context, status Status, limit, offset int) ([]Order, error)
	SetPayment(ctx context.Context, orderID int64, paymentID, paymentURL string) error
	// ChangeStatus меняет статус, только если заказ всё ещё в статусе from, и сообщает,
	// изменилась ли строка. Так повтор (второе уведомление ЮKassa, двойной клик)
	// ничего не меняет второй раз.
	ChangeStatus(ctx context.Context, orderID int64, from, to Status) (bool, error)
}
