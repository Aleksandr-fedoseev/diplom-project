// Package shop — магазин сувениров: каталог товаров, корзина, заказы и оплата через ЮKassa.
// Модуль — владелец таблиц products, cart_items, orders и order_items. Сведения о
// пользователях он получает у модуля auth, письма отправляет через модуль notifications.
//
// Цены в API — целые копейки (priceKopecks, totalKopecks), как и в БД:
// 4500 ₽ передаётся как 450000. Форматирует сумму для показа фронтенд.
package shop

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/auth"
	"heritage-backend/internal/httpx"
	"heritage-backend/internal/notifications"
	"heritage-backend/internal/shop/store"
)

// Статусы заказа.
const (
	StatusPending   = "pending"   // ждёт оплаты
	StatusPaid      = "paid"      // оплачен
	StatusShipped   = "shipped"   // отправлен
	StatusCompleted = "completed" // выполнен
	StatusCancelled = "cancelled" // отменён
)

// transitions — какие смены статуса может сделать администратор.
// Оплата (pending → paid) сюда не входит: её подтверждает только ЮKassa.
// Отмена оплаченного заказа не возвращает деньги автоматически: возврат
// делается вручную в личном кабинете ЮKassa (автоматические возвраты — вне рамок проекта).
var transitions = map[string][]string{
	StatusPending: {StatusCancelled},
	StatusPaid:    {StatusShipped, StatusCancelled},
	StatusShipped: {StatusCompleted},
}

var validStatuses = map[string]bool{
	StatusPending: true, StatusPaid: true, StatusShipped: true, StatusCompleted: true, StatusCancelled: true,
}

type Product struct {
	ID           int64     `json:"id"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Category     string    `json:"category"`
	PriceKopecks int64     `json:"priceKopecks"`
	Stock        int32     `json:"stock"`
	InStock      bool      `json:"inStock"`
	ImageURL     string    `json:"imageUrl"`
	IsActive     bool      `json:"isActive"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func productFromRow(p store.Product) Product {
	return Product{
		ID:           p.ID,
		Slug:         p.Slug,
		Title:        p.Title,
		Description:  p.Description,
		Category:     p.Category,
		PriceKopecks: p.PriceKopecks,
		Stock:        p.Stock,
		InStock:      p.IsActive && p.Stock > 0,
		ImageURL:     p.ImageUrl,
		IsActive:     p.IsActive,
		UpdatedAt:    p.UpdatedAt,
	}
}

type CartItem struct {
	ProductID        int64  `json:"productId"`
	Slug             string `json:"slug"`
	Title            string `json:"title"`
	ImageURL         string `json:"imageUrl"`
	PriceKopecks     int64  `json:"priceKopecks"`
	Quantity         int32  `json:"quantity"`
	Stock            int32  `json:"stock"`
	LineTotalKopecks int64  `json:"lineTotalKopecks"`
	// Available — товар в продаже и его хватает; иначе оформить заказ не получится.
	Available bool `json:"available"`
}

type Cart struct {
	Items        []CartItem `json:"items"`
	TotalKopecks int64      `json:"totalKopecks"`
}

type OrderItem struct {
	ProductID        int64  `json:"productId"`
	Title            string `json:"title"`
	Quantity         int32  `json:"quantity"`
	PriceKopecks     int64  `json:"priceKopecks"`
	LineTotalKopecks int64  `json:"lineTotalKopecks"`
}

type Order struct {
	ID              int64       `json:"id"`
	UserID          *int64      `json:"userId,omitempty"`
	Status          string      `json:"status"`
	TotalKopecks    int64       `json:"totalKopecks"`
	PaymentURL      string      `json:"paymentUrl"` // страница оплаты ЮKassa; пусто, если платёж не создан
	DeliveryAddress string      `json:"deliveryAddress"`
	Phone           string      `json:"phone"`
	Items           []OrderItem `json:"items"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
}

func orderFromRow(o store.Order) Order {
	return Order{
		ID:              o.ID,
		UserID:          o.UserID,
		Status:          o.Status,
		TotalKopecks:    o.TotalKopecks,
		PaymentURL:      o.PaymentUrl,
		DeliveryAddress: o.DeliveryAddress,
		Phone:           o.Phone,
		Items:           []OrderItem{},
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
}

// Users — то, что магазину нужно от модуля auth. Реализует auth.Service.
type Users interface {
	GetByID(ctx context.Context, id int64) (auth.User, error)
	EnsureActive(ctx context.Context, userID int64) error
}

// Notifier — очередь писем. Реализует notifications.Service.
type Notifier interface {
	Enqueue(ctx context.Context, msg notifications.Message) error
	EnqueueTx(ctx context.Context, tx pgx.Tx, msg notifications.Message) error
}

type Service struct {
	db        *pgxpool.Pool  // нужен только для транзакций
	q         *store.Queries // запросы из queries.sql, код сгенерирован sqlc
	users     Users
	notify    Notifier
	payments  PaymentGateway
	returnURL string
}

// NewService создаёт сервис магазина. payments может быть nil — тогда, как и с
// ненастроенной ЮKassa, заказы оформляются без создания платежа.
// returnURL — страница сайта, куда ЮKassa вернёт покупателя; к ней добавляется ?order=<id>.
func NewService(db *pgxpool.Pool, users Users, notify Notifier, payments PaymentGateway, returnURL string) *Service {
	return &Service{db: db, q: store.New(db), users: users, notify: notify, payments: payments, returnURL: returnURL}
}

// Товары.

// ListProducts — товары в продаже, по желанию одной категории. includeInactive — и снятые (для администратора).
func (s *Service) ListProducts(ctx context.Context, category string, includeInactive bool) ([]Product, error) {
	var cat *string
	if category != "" {
		cat = &category
	}
	rows, err := s.q.ListProducts(ctx, store.ListProductsParams{IncludeInactive: includeInactive, Category: cat})
	if err != nil {
		return nil, err
	}
	list := make([]Product, len(rows))
	for i, p := range rows {
		list[i] = productFromRow(p)
	}
	return list, nil
}

// GetProduct возвращает карточку товара. Снятый с продажи товар видит только администратор.
func (s *Service) GetProduct(ctx context.Context, slug string, includeInactive bool) (Product, error) {
	p, err := s.q.GetProductBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !p.IsActive && !includeInactive) {
		return Product{}, httpx.NotFound("товар не найден")
	}
	if err != nil {
		return Product{}, err
	}
	return productFromRow(p), nil
}

// ProductInput — данные для создания и изменения товара администратором.
type ProductInput struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	PriceKopecks int64  `json:"priceKopecks"`
	Stock        int32  `json:"stock"`
	ImageURL     string `json:"imageUrl"`
	IsActive     *bool  `json:"isActive"` // по умолчанию true
}

func (in *ProductInput) validate() error {
	in.Slug = strings.TrimSpace(strings.ToLower(in.Slug))
	in.Title = strings.TrimSpace(in.Title)
	in.Category = strings.TrimSpace(in.Category)
	switch {
	case in.Slug == "" || strings.ContainsAny(in.Slug, " /?#"):
		return httpx.BadRequest("slug обязателен и не должен содержать пробелы и / ? #")
	case in.Title == "":
		return httpx.BadRequest("укажите название")
	case in.PriceKopecks <= 0:
		return httpx.BadRequest("priceKopecks — цена в копейках, больше нуля")
	case in.Stock < 0:
		return httpx.BadRequest("stock не может быть отрицательным")
	}
	if in.IsActive == nil {
		t := true
		in.IsActive = &t
	}
	return nil
}

func (s *Service) CreateProduct(ctx context.Context, in ProductInput) (Product, error) {
	if err := in.validate(); err != nil {
		return Product{}, err
	}
	p, err := s.q.CreateProduct(ctx, store.CreateProductParams{
		Slug:         in.Slug,
		Title:        in.Title,
		Description:  in.Description,
		Category:     in.Category,
		PriceKopecks: in.PriceKopecks,
		Stock:        in.Stock,
		ImageUrl:     in.ImageURL,
		IsActive:     *in.IsActive,
	})
	if err := mapWriteError(err); err != nil {
		return Product{}, err
	}
	return productFromRow(p), nil
}

// UpdateProduct меняет товар. Новая цена действует для новых заказов:
// в оформленных заказах остаётся цена на момент покупки.
func (s *Service) UpdateProduct(ctx context.Context, id int64, in ProductInput) (Product, error) {
	if err := in.validate(); err != nil {
		return Product{}, err
	}
	p, err := s.q.UpdateProduct(ctx, store.UpdateProductParams{
		ID:           id,
		Slug:         in.Slug,
		Title:        in.Title,
		Description:  in.Description,
		Category:     in.Category,
		PriceKopecks: in.PriceKopecks,
		Stock:        in.Stock,
		ImageUrl:     in.ImageURL,
		IsActive:     *in.IsActive,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, httpx.NotFound("товар не найден")
	}
	if err := mapWriteError(err); err != nil {
		return Product{}, err
	}
	return productFromRow(p), nil
}

// DeactivateProduct снимает товар с продажи. Физически товар не удаляется:
// на него ссылаются позиции оформленных заказов, и история покупок должна сохраниться.
// Из корзин товар убирается.
func (s *Service) DeactivateProduct(ctx context.Context, id int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }() // после Commit ничего не делает

	qtx := s.q.WithTx(tx)
	n, err := qtx.DeactivateProduct(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.NotFound("товар не найден")
	}
	if err := qtx.DeleteProductFromCarts(ctx, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// mapWriteError превращает ошибки ограничений БД в понятные клиенту ответы.
func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return httpx.Conflict("товар с таким slug уже есть")
	}
	return err
}

// Корзина.

func (s *Service) GetCart(ctx context.Context, userID int64) (Cart, error) {
	rows, err := s.q.ListCartItems(ctx, userID)
	if err != nil {
		return Cart{}, err
	}
	cart := Cart{Items: make([]CartItem, len(rows))}
	for i, r := range rows {
		item := CartItem{
			ProductID:        r.ProductID,
			Slug:             r.Slug,
			Title:            r.Title,
			ImageURL:         r.ImageUrl,
			PriceKopecks:     r.PriceKopecks,
			Quantity:         r.Quantity,
			Stock:            r.Stock,
			LineTotalKopecks: r.PriceKopecks * int64(r.Quantity),
			Available:        r.IsActive && r.Quantity <= r.Stock,
		}
		cart.Items[i] = item
		cart.TotalKopecks += item.LineTotalKopecks
	}
	return cart, nil
}

// SetCartItem задаёт количество товара в корзине; 0 убирает товар.
// Больше, чем есть на складе, положить нельзя. Окончательно наличие
// проверяется при оформлении заказа: пока товар лежит в корзине, его могут купить другие.
func (s *Service) SetCartItem(ctx context.Context, userID, productID int64, quantity int32) (Cart, error) {
	if quantity < 0 {
		return Cart{}, httpx.BadRequest("quantity не может быть отрицательным")
	}
	if quantity == 0 {
		return s.RemoveCartItem(ctx, userID, productID)
	}
	p, err := s.q.GetProductByID(ctx, productID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !p.IsActive) {
		return Cart{}, httpx.NotFound("товар не найден")
	}
	if err != nil {
		return Cart{}, err
	}
	if quantity > p.Stock {
		return Cart{}, httpx.Conflict(fmt.Sprintf("в наличии только %d шт. товара «%s»", p.Stock, p.Title))
	}
	err = s.q.UpsertCartItem(ctx, store.UpsertCartItemParams{UserID: userID, ProductID: productID, Quantity: quantity})
	if err != nil {
		return Cart{}, err
	}
	return s.GetCart(ctx, userID)
}

func (s *Service) RemoveCartItem(ctx context.Context, userID, productID int64) (Cart, error) {
	if err := s.q.DeleteCartItem(ctx, store.DeleteCartItemParams{UserID: userID, ProductID: productID}); err != nil {
		return Cart{}, err
	}
	return s.GetCart(ctx, userID)
}

// Оформление заказа и оплата.

type CheckoutInput struct {
	DeliveryAddress string `json:"deliveryAddress"`
	Phone           string `json:"phone"`
}

func (in *CheckoutInput) validate() error {
	in.DeliveryAddress = strings.TrimSpace(in.DeliveryAddress)
	in.Phone = strings.TrimSpace(in.Phone)
	if in.DeliveryAddress == "" || len([]rune(in.DeliveryAddress)) > 500 {
		return httpx.BadRequest("укажите адрес доставки (до 500 символов)")
	}
	// Телефон: цифры, пробелы, скобки, дефисы и «+» в начале; цифр — от 10 до 15.
	digits := 0
	for i, r := range in.Phone {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == ' ' || r == '(' || r == ')' || r == '-' || (r == '+' && i == 0):
		default:
			return httpx.BadRequest("некорректный номер телефона")
		}
	}
	if digits < 10 || digits > 15 {
		return httpx.BadRequest("некорректный номер телефона")
	}
	return nil
}

// Checkout оформляет заказ из корзины и создаёт платёж в ЮKassa.
//
// Всё, что касается склада, делается в одной транзакции: строки товаров блокируются,
// наличие проверяется, остатки уменьшаются, заказ с позициями записывается, корзина
// очищается. Сумма считается на сервере по текущим ценам, в позициях заказа
// запоминаются цена и название на момент покупки.
//
// Платёж создаётся после фиксации транзакции: держать блокировки во время запроса
// к внешнему сервису нельзя. Если ЮKassa не ответила, заказ всё равно сохранён,
// а ссылку на оплату можно получить позже через Pay.
func (s *Service) Checkout(ctx context.Context, userID int64, in CheckoutInput) (Order, error) {
	if err := in.validate(); err != nil {
		return Order{}, err
	}
	if err := s.users.EnsureActive(ctx, userID); err != nil {
		return Order{}, err
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return Order{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }() // после Commit ничего не делает

	qtx := s.q.WithTx(tx)
	lines, err := qtx.LockCartProducts(ctx, userID)
	if err != nil {
		return Order{}, err
	}
	if len(lines) == 0 {
		return Order{}, httpx.BadRequest("корзина пуста")
	}

	var total int64
	for _, l := range lines {
		if !l.IsActive {
			return Order{}, httpx.Conflict(fmt.Sprintf("товар «%s» снят с продажи — уберите его из корзины", l.Title))
		}
		if l.Quantity > l.Stock {
			return Order{}, httpx.Conflict(fmt.Sprintf("товара «%s» осталось %d шт.", l.Title, l.Stock))
		}
		total += l.PriceKopecks * int64(l.Quantity)
	}

	orderID, err := qtx.CreateOrder(ctx, store.CreateOrderParams{
		UserID:          &userID,
		TotalKopecks:    total,
		DeliveryAddress: in.DeliveryAddress,
		Phone:           in.Phone,
	})
	if err != nil {
		return Order{}, err
	}
	for _, l := range lines {
		err := qtx.CreateOrderItem(ctx, store.CreateOrderItemParams{
			OrderID: orderID, ProductID: l.ProductID, Title: l.Title, Quantity: l.Quantity, PriceKopecks: l.PriceKopecks,
		})
		if err != nil {
			return Order{}, err
		}
		if err := qtx.DecrementStock(ctx, store.DecrementStockParams{ID: l.ProductID, Quantity: l.Quantity}); err != nil {
			return Order{}, err
		}
	}
	if err := qtx.ClearCart(ctx, userID); err != nil {
		return Order{}, err
	}

	var body strings.Builder
	fmt.Fprintf(&body, "Здравствуйте, %s!\n\nВаш заказ №%d оформлен.\n\n", user.Name, orderID)
	for _, l := range lines {
		fmt.Fprintf(&body, "%s — %d шт. × %s ₽\n", l.Title, l.Quantity, formatAmount(l.PriceKopecks))
	}
	fmt.Fprintf(&body, "\nИтого: %s ₽\nАдрес доставки: %s\n\nЗаказ будет передан в доставку после оплаты.\n",
		formatAmount(total), in.DeliveryAddress)
	err = s.notify.EnqueueTx(ctx, tx, notifications.Message{
		To: user.Email, Subject: fmt.Sprintf("Заказ №%d оформлен", orderID), Body: body.String(),
	})
	if err != nil {
		return Order{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, err
	}

	if err := s.createPayment(ctx, orderID, total); err != nil {
		if errors.Is(err, ErrPaymentsDisabled) {
			slog.Warn("ЮKassa не настроена: заказ оформлен без платежа", "order", orderID)
		} else {
			slog.Error("платёж не создан, покупатель может повторить оплату позже", "order", orderID, "err", err)
		}
	}
	return s.getOrder(ctx, orderID)
}

// Pay возвращает заказ со ссылкой на оплату, создавая платёж, если его ещё нет
// (например, при оформлении ЮKassa была недоступна).
func (s *Service) Pay(ctx context.Context, orderID, userID int64) (Order, error) {
	o, err := s.q.GetOrder(ctx, orderID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && (o.UserID == nil || *o.UserID != userID)) {
		return Order{}, httpx.NotFound("заказ не найден")
	}
	if err != nil {
		return Order{}, err
	}
	if o.Status != StatusPending {
		return Order{}, httpx.Conflict("заказ уже оплачен или отменён")
	}
	if o.PaymentUrl == "" {
		err := s.createPayment(ctx, o.ID, o.TotalKopecks)
		if errors.Is(err, ErrPaymentsDisabled) {
			return Order{}, &httpx.Error{Status: http.StatusServiceUnavailable, Message: "приём оплаты временно не работает"}
		}
		if err != nil {
			slog.Error("платёж не создан", "order", o.ID, "err", err)
			return Order{}, &httpx.Error{Status: http.StatusBadGateway, Message: "платёжный сервис недоступен, попробуйте позже"}
		}
	}
	return s.getOrder(ctx, orderID)
}

// createPayment создаёт платёж в ЮKassa и запоминает его id и страницу оплаты в заказе.
func (s *Service) createPayment(ctx context.Context, orderID, totalKopecks int64) error {
	if s.payments == nil {
		return ErrPaymentsDisabled
	}
	p, err := s.payments.CreatePayment(ctx, PaymentRequest{
		OrderID:       orderID,
		AmountKopecks: totalKopecks,
		Description:   fmt.Sprintf("Заказ №%d", orderID),
		ReturnURL:     s.orderReturnURL(orderID),
	})
	if err != nil {
		return err
	}
	return s.q.SetOrderPayment(ctx, store.SetOrderPaymentParams{ID: orderID, PaymentID: &p.ID, PaymentUrl: p.ConfirmationURL})
}

// orderReturnURL добавляет к адресу возврата номер заказа, чтобы сайт показал нужный заказ.
func (s *Service) orderReturnURL(orderID int64) string {
	u, err := url.Parse(s.returnURL)
	if err != nil {
		return s.returnURL
	}
	q := u.Query()
	q.Set("order", strconv.FormatInt(orderID, 10))
	u.RawQuery = q.Encode()
	return u.String()
}

// Уведомления ЮKassa.

// HandlePaymentNotification обрабатывает уведомление ЮKassa о платеже paymentID.
//
// Уведомления ЮKassa не подписаны, поэтому телу запроса не доверяем: берём из него
// только id платежа и сами запрашиваем платёж у ЮKassa по API — действуем по статусу из ответа.
// Дополнительно на nginx можно пропускать к адресу уведомлений только IP-адреса ЮKassa
// (список есть в документации: https://yookassa.ru/developers/using-api/webhooks).
//
// Обработка идемпотентна: статус меняется условным UPDATE ... WHERE status = 'pending',
// и письмо или возврат товаров на склад делаются, только если строка действительно изменилась.
// Ошибка возвращается только для временных сбоев — тогда ЮKassa повторит уведомление.
// Неизвестный платёж ошибкой не считается, чтобы ЮKassa перестала его присылать.
func (s *Service) HandlePaymentNotification(ctx context.Context, paymentID string) error {
	if s.payments == nil {
		slog.Warn("уведомление ЮKassa, но приём оплаты не настроен", "payment", paymentID)
		return nil
	}
	p, err := s.payments.GetPayment(ctx, paymentID)
	if errors.Is(err, ErrPaymentNotFound) || errors.Is(err, ErrPaymentsDisabled) {
		slog.Warn("уведомление о неизвестном платеже", "payment", paymentID, "err", err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("запрос платежа %s: %w", paymentID, err)
	}

	o, err := s.q.GetOrderByPaymentID(ctx, &p.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		slog.Warn("платёж не привязан ни к одному заказу", "payment", p.ID, "metadataOrder", p.OrderID)
		return nil
	}
	if err != nil {
		return err
	}

	switch p.Status {
	case PaymentSucceeded:
		if p.Amount != formatAmount(o.TotalKopecks) {
			slog.Error("сумма платежа не совпадает с суммой заказа", "order", o.ID, "payment", p.ID, "amount", p.Amount)
			return nil
		}
		changed, err := s.changeStatus(ctx, o, StatusPending, StatusPaid)
		if err != nil {
			return err
		}
		if !changed && o.Status == StatusCancelled {
			// Покупатель оплатил уже отменённый заказ — деньги нужно вернуть вручную в кабинете ЮKassa.
			slog.Warn("оплачен отменённый заказ, нужен возврат", "order", o.ID, "payment", p.ID)
		}
	case PaymentCanceled:
		if _, err := s.changeStatus(ctx, o, StatusPending, StatusCancelled); err != nil {
			return err
		}
	default:
		// pending и waiting_for_capture: платёж ещё не завершён, ждём следующего уведомления.
	}
	return nil
}

// Заказы.

// GetOrder возвращает заказ с позициями. Чужой заказ видит только администратор.
func (s *Service) GetOrder(ctx context.Context, id int64, p auth.Principal) (Order, error) {
	o, err := s.getOrder(ctx, id)
	if err != nil {
		return Order{}, err
	}
	if p.Role != auth.RoleAdmin && (o.UserID == nil || *o.UserID != p.UserID) {
		return Order{}, httpx.NotFound("заказ не найден")
	}
	return o, nil
}

func (s *Service) getOrder(ctx context.Context, id int64) (Order, error) {
	row, err := s.q.GetOrder(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, httpx.NotFound("заказ не найден")
	}
	if err != nil {
		return Order{}, err
	}
	list := []Order{orderFromRow(row)}
	if err := s.fillItems(ctx, list); err != nil {
		return Order{}, err
	}
	return list[0], nil
}

func (s *Service) ListMyOrders(ctx context.Context, userID int64) ([]Order, error) {
	rows, err := s.q.ListOrdersByUser(ctx, &userID)
	if err != nil {
		return nil, err
	}
	list := make([]Order, len(rows))
	for i, r := range rows {
		list[i] = orderFromRow(r)
	}
	return list, s.fillItems(ctx, list)
}

// ListOrders — все заказы для администратора, по желанию с одним статусом.
func (s *Service) ListOrders(ctx context.Context, status string, limit, offset int32) ([]Order, error) {
	var st *string
	if status != "" {
		if !validStatuses[status] {
			return nil, httpx.BadRequest("status: pending, paid, shipped, completed или cancelled")
		}
		st = &status
	}
	rows, err := s.q.ListOrders(ctx, store.ListOrdersParams{Status: st, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	list := make([]Order, len(rows))
	for i, r := range rows {
		list[i] = orderFromRow(r)
	}
	return list, s.fillItems(ctx, list)
}

// fillItems заполняет позиции сразу для всех заказов списка одним запросом.
func (s *Service) fillItems(ctx context.Context, list []Order) error {
	if len(list) == 0 {
		return nil
	}
	ids := make([]int64, len(list))
	for i, o := range list {
		ids[i] = o.ID
	}
	rows, err := s.q.ListOrderItems(ctx, ids)
	if err != nil {
		return err
	}
	items := make(map[int64][]OrderItem, len(list))
	for _, r := range rows {
		items[r.OrderID] = append(items[r.OrderID], OrderItem{
			ProductID:        r.ProductID,
			Title:            r.Title,
			Quantity:         r.Quantity,
			PriceKopecks:     r.PriceKopecks,
			LineTotalKopecks: r.PriceKopecks * int64(r.Quantity),
		})
	}
	for i := range list {
		if it := items[list[i].ID]; it != nil {
			list[i].Items = it
		}
	}
	return nil
}

// CancelMine — покупатель отменяет свой неоплаченный заказ, товары возвращаются на склад.
// Без этого брошенный заказ держал бы товар, пока платёж не истечёт в ЮKassa.
func (s *Service) CancelMine(ctx context.Context, orderID, userID int64) (Order, error) {
	o, err := s.q.GetOrder(ctx, orderID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && (o.UserID == nil || *o.UserID != userID)) {
		return Order{}, httpx.NotFound("заказ не найден")
	}
	if err != nil {
		return Order{}, err
	}
	changed, err := s.changeStatus(ctx, o, StatusPending, StatusCancelled)
	if err != nil {
		return Order{}, err
	}
	if !changed {
		return Order{}, httpx.Conflict("отменить можно только неоплаченный заказ")
	}
	return s.getOrder(ctx, orderID)
}

// SetStatus — смена статуса заказа администратором (только разрешённые переходы, см. transitions).
func (s *Service) SetStatus(ctx context.Context, id int64, status string) (Order, error) {
	if !validStatuses[status] {
		return Order{}, httpx.BadRequest("status: pending, paid, shipped, completed или cancelled")
	}
	o, err := s.q.GetOrder(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, httpx.NotFound("заказ не найден")
	}
	if err != nil {
		return Order{}, err
	}
	allowed := false
	for _, next := range transitions[o.Status] {
		allowed = allowed || next == status
	}
	if !allowed {
		return Order{}, httpx.Conflict(fmt.Sprintf("нельзя перевести заказ из статуса %s в %s", o.Status, status))
	}
	changed, err := s.changeStatus(ctx, o, o.Status, status)
	if err != nil {
		return Order{}, err
	}
	if !changed {
		return Order{}, httpx.Conflict("статус заказа только что изменился, обновите страницу")
	}
	return s.getOrder(ctx, id)
}

// changeStatus переводит заказ из статуса from в status. Если заказ уже не в статусе from
// (повторное уведомление, параллельный запрос), ничего не делает и возвращает false.
// Сведения из o (с прочитанным ранее статусом) нужны только для письма. При отмене товары возвращаются на склад. Письмо покупателю
// ставится в очередь в той же транзакции.
func (s *Service) changeStatus(ctx context.Context, o store.Order, from, status string) (bool, error) {
	msg, hasMsg, err := s.statusEmail(ctx, o, from, status)
	if err != nil {
		return false, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }() // после Commit ничего не делает

	qtx := s.q.WithTx(tx)
	n, err := qtx.ChangeOrderStatus(ctx, store.ChangeOrderStatusParams{ID: o.ID, OldStatus: from, NewStatus: status})
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, nil
	}
	if status == StatusCancelled {
		if err := qtx.ReturnOrderStock(ctx, o.ID); err != nil {
			return false, err
		}
	}
	if hasMsg {
		if err := s.notify.EnqueueTx(ctx, tx, msg); err != nil {
			return false, err
		}
	}
	return true, tx.Commit(ctx)
}

// statusEmail готовит письмо покупателю о новом статусе заказа.
// Если покупатель удалён или о статусе писать незачем, hasMsg = false.
func (s *Service) statusEmail(ctx context.Context, o store.Order, from, status string) (msg notifications.Message, hasMsg bool, err error) {
	var subject, text string
	switch status {
	case StatusPaid:
		subject = fmt.Sprintf("Оплата заказа №%d получена", o.ID)
		text = fmt.Sprintf("Мы получили оплату заказа №%d на сумму %s ₽. Сообщим, когда заказ будет отправлен.",
			o.ID, formatAmount(o.TotalKopecks))
	case StatusShipped:
		subject = fmt.Sprintf("Заказ №%d отправлен", o.ID)
		text = fmt.Sprintf("Заказ №%d передан в доставку по адресу: %s.", o.ID, o.DeliveryAddress)
	case StatusCancelled:
		subject = fmt.Sprintf("Заказ №%d отменён", o.ID)
		text = fmt.Sprintf("Заказ №%d отменён.", o.ID)
		if from == StatusPaid {
			text += " Деньги будут возвращены на карту, с которой производилась оплата."
		}
	default:
		return msg, false, nil
	}
	if o.UserID == nil {
		return msg, false, nil
	}
	u, err := s.users.GetByID(ctx, *o.UserID)
	var he *httpx.Error
	if errors.As(err, &he) && he.Status == http.StatusNotFound {
		return msg, false, nil
	}
	if err != nil {
		return msg, false, err
	}
	return notifications.Message{
		To:      u.Email,
		Subject: subject,
		Body:    fmt.Sprintf("Здравствуйте, %s!\n\n%s\n", u.Name, text),
	}, true, nil
}
