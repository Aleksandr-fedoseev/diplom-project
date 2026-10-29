package shop

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"heritage-backend/internal/domain/errs"
	"heritage-backend/internal/domain/shop"
)

// Тесты сервиса без БД и HTTP: хранилища, транзакции, события и ЮKassa — фейки в памяти.

type env struct {
	t     *testing.T
	ctx   context.Context
	db    *memDB
	svc   *Service
	gw    *fakeGateway
	users fakeUsers
}

func newEnv(t *testing.T) *env {
	db := newMemDB()
	gw := newFakeGateway()
	users := fakeUsers{blocked: map[int64]bool{}}
	svc := NewService(Deps{
		Tx: fakeTx{db}, Products: fakeProducts{db}, Carts: fakeCarts{db}, Orders: fakeOrders{db},
		Users: users, Events: fakeEvents{db}, Payments: gw, ReturnURL: "https://site.test/shop/orders",
	})
	return &env{t: t, ctx: context.Background(), db: db, svc: svc, gw: gw, users: users}
}

var nextUser int64 = 1000

func newUserID() int64 { nextUser++; return nextUser }

func (e *env) newProduct(price int64, stock int) shop.Product {
	e.t.Helper()
	p, err := e.svc.CreateProduct(e.ctx, shop.ProductInput{
		Slug: fmt.Sprintf("p-%d", e.db.nextID+1), Title: fmt.Sprintf("Товар %d", e.db.nextID+1), Category: "test",
		PriceKopecks: price, Stock: stock,
	})
	if err != nil {
		e.t.Fatal(err)
	}
	return p
}

func (e *env) stock(id int64) int { return e.db.products[id].Stock }

func (e *env) put(userID, productID int64, qty int) shop.Cart {
	e.t.Helper()
	cart, err := e.svc.SetCartItem(e.ctx, userID, productID, qty)
	if err != nil {
		e.t.Fatal(err)
	}
	return cart
}

// eventsOf — события указанного типа.
func (e *env) eventsOf(typ string) []any {
	var out []any
	for _, ev := range e.db.events {
		if ev.Type == typ {
			out = append(out, ev.Payload)
		}
	}
	return out
}

func wantKind(t *testing.T, err error, kind errs.Kind) {
	t.Helper()
	if !errs.Is(err, kind) {
		t.Fatalf("ожидалась ошибка рода %d, получено: %v", kind, err)
	}
}

var validCheckout = shop.CheckoutInput{DeliveryAddress: "г. Астрахань, ул. Никольская, 1", Phone: "+7 (900) 123-45-67"}

func TestProducts(t *testing.T) {
	e := newEnv(t)
	p := e.newProduct(99900, 3)

	_, err := e.svc.CreateProduct(e.ctx, shop.ProductInput{Slug: p.Slug, Title: "Дубликат", PriceKopecks: 100})
	wantKind(t, err, errs.Conflict)
	_, err = e.svc.CreateProduct(e.ctx, shop.ProductInput{Slug: "free", Title: "Бесплатно", PriceKopecks: 0})
	wantKind(t, err, errs.Invalid)

	got, err := e.svc.GetProduct(e.ctx, p.Slug, false)
	if err != nil || got.PriceKopecks != 99900 || !got.InStock() {
		t.Fatalf("карточка товара: %+v, %v", got, err)
	}

	// Снятый с продажи товар пропадает с витрины и из корзин, но остаётся в БД.
	u := newUserID()
	e.put(u, p.ID, 1)
	if err := e.svc.DeactivateProduct(e.ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.GetProduct(e.ctx, p.Slug, false)
	wantKind(t, err, errs.NotFound)
	if _, err := e.svc.GetProduct(e.ctx, p.Slug, true); err != nil {
		t.Fatalf("администратор должен видеть снятый товар: %v", err)
	}
	if list, _ := e.svc.ListProducts(e.ctx, "test", false); len(list) != 0 {
		t.Fatalf("снятый товар не должен быть на витрине: %+v", list)
	}
	if cart, _ := e.svc.GetCart(e.ctx, u); len(cart.Items) != 0 {
		t.Fatalf("снятый товар должен пропасть из корзины: %+v", cart)
	}
	wantKind(t, e.svc.DeactivateProduct(e.ctx, -1), errs.NotFound)
	_, err = e.svc.UpdateProduct(e.ctx, -1, shop.ProductInput{Slug: "x", Title: "x", PriceKopecks: 1})
	wantKind(t, err, errs.NotFound)
}

func TestCart(t *testing.T) {
	e := newEnv(t)
	u := newUserID()
	a := e.newProduct(450000, 5)
	b := e.newProduct(120000, 2)

	e.put(u, a.ID, 2)
	cart := e.put(u, b.ID, 1)
	if len(cart.Items) != 2 || cart.Total() != 2*450000+120000 {
		t.Fatalf("корзина: %+v", cart)
	}
	if cart.Items[0].ProductID != a.ID || cart.Items[0].LineTotal() != 900000 || !cart.Items[0].Available() {
		t.Fatalf("позиция корзины: %+v", cart.Items[0])
	}

	// Повторный PUT задаёт количество, а не прибавляет.
	cart = e.put(u, a.ID, 5)
	if cart.Items[0].Quantity != 5 || cart.Total() != 5*450000+120000 {
		t.Fatalf("после изменения количества: %+v", cart)
	}

	_, err := e.svc.SetCartItem(e.ctx, u, a.ID, 6)
	wantKind(t, err, errs.Conflict) // больше, чем на складе
	_, err = e.svc.SetCartItem(e.ctx, u, a.ID, -1)
	wantKind(t, err, errs.Invalid)
	_, err = e.svc.SetCartItem(e.ctx, u, -1, 1)
	wantKind(t, err, errs.NotFound)

	cart = e.put(u, b.ID, 0) // 0 убирает товар
	if len(cart.Items) != 1 || cart.Total() != 5*450000 {
		t.Fatalf("после удаления количеством 0: %+v", cart)
	}
	cart, err = e.svc.RemoveCartItem(e.ctx, u, a.ID)
	if err != nil || len(cart.Items) != 0 || cart.Total() != 0 {
		t.Fatalf("пустая корзина: %+v, %v", cart, err)
	}
}

func TestCheckout(t *testing.T) {
	e := newEnv(t)
	u := newUserID()
	a := e.newProduct(450000, 5)
	b := e.newProduct(90000, 10)
	e.put(u, b.ID, 3)
	e.put(u, a.ID, 2)

	o, err := e.svc.Checkout(e.ctx, u, validCheckout)
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != shop.StatusPending || o.TotalKopecks != 2*450000+3*90000 || len(o.Items) != 2 ||
		o.DeliveryAddress != validCheckout.DeliveryAddress || o.Phone != validCheckout.Phone || !o.OwnedBy(u) {
		t.Fatalf("заказ: %+v", o)
	}
	if o.PaymentURL != fmt.Sprintf("https://yoomoney.test/checkout?order=%d", o.ID) {
		t.Fatalf("ссылка на оплату: %q", o.PaymentURL)
	}
	req := e.gw.created[0]
	if req.Amount != o.TotalKopecks || req.ReturnURL != fmt.Sprintf("https://site.test/shop/orders?order=%d", o.ID) ||
		req.Description != fmt.Sprintf("Заказ №%d", o.ID) || req.IdempotenceKey != fmt.Sprintf("order-%d", o.ID) {
		t.Fatalf("запрос платежа: %+v", req)
	}
	if e.stock(a.ID) != 3 || e.stock(b.ID) != 7 {
		t.Fatalf("остатки после заказа: %d, %d", e.stock(a.ID), e.stock(b.ID))
	}
	if cart, _ := e.svc.GetCart(e.ctx, u); len(cart.Items) != 0 {
		t.Fatalf("корзина после заказа должна быть пустой: %+v", cart)
	}
	placed := e.eventsOf(EventOrderPlaced)
	wantPlaced := OrderPlaced{OrderID: o.ID, UserID: u, TotalKopecks: o.TotalKopecks, DeliveryAddress: validCheckout.DeliveryAddress,
		Items: []OrderPlacedItem{{Title: a.Title, Quantity: 2, PriceKopecks: 450000}, {Title: b.Title, Quantity: 3, PriceKopecks: 90000}}}
	if len(placed) != 1 || !reflect.DeepEqual(placed[0], wantPlaced) {
		t.Fatalf("событие order.placed: %+v", placed)
	}

	// Позиции хранят цену и название на момент покупки: смена цены их не меняет.
	if _, err := e.svc.UpdateProduct(e.ctx, a.ID, shop.ProductInput{Slug: a.Slug, Title: "Новое название", PriceKopecks: 1, Stock: 3}); err != nil {
		t.Fatal(err)
	}
	got, err := e.svc.GetOrder(e.ctx, o.ID, shop.Viewer{UserID: u})
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range got.Items {
		if it.ProductID == a.ID && (it.PriceKopecks != 450000 || it.Title != a.Title || it.LineTotal() != 900000) {
			t.Fatalf("снимок позиции: %+v", it)
		}
	}

	// Чужой заказ не виден, администратору виден.
	_, err = e.svc.GetOrder(e.ctx, o.ID, shop.Viewer{UserID: newUserID()})
	wantKind(t, err, errs.NotFound)
	if _, err := e.svc.GetOrder(e.ctx, o.ID, shop.Viewer{UserID: newUserID(), IsAdmin: true}); err != nil {
		t.Fatal(err)
	}
	if mine, err := e.svc.ListMyOrders(e.ctx, u); err != nil || len(mine) != 1 || mine[0].ID != o.ID {
		t.Fatalf("мои заказы: %+v, %v", mine, err)
	}

	// Pay для заказа с платежом не создаёт второй платёж.
	if _, err := e.svc.Pay(e.ctx, o.ID, u); err != nil || len(e.gw.created) != 1 {
		t.Fatalf("повторная оплата: %v, платежей %d", err, len(e.gw.created))
	}
}

func TestCheckoutInsufficientStock(t *testing.T) {
	e := newEnv(t)
	u := newUserID()
	a := e.newProduct(10000, 5)
	b := e.newProduct(20000, 2)
	e.put(u, a.ID, 3)
	e.put(u, b.ID, 2)

	// Пока товар лежал в корзине, остаток уменьшился (купили другие).
	if _, err := e.svc.UpdateProduct(e.ctx, b.ID, shop.ProductInput{Slug: b.Slug, Title: b.Title, PriceKopecks: 20000, Stock: 1}); err != nil {
		t.Fatal(err)
	}
	if cart, _ := e.svc.GetCart(e.ctx, u); cart.Items[1].Available() {
		t.Fatalf("позиция без достаточного остатка должна быть недоступна: %+v", cart.Items[1])
	}

	_, err := e.svc.Checkout(e.ctx, u, validCheckout)
	wantKind(t, err, errs.Conflict)

	// Ничего не изменилось: остатки, корзина, заказов и событий нет.
	if e.stock(a.ID) != 5 || e.stock(b.ID) != 1 {
		t.Fatalf("остатки не должны меняться: %d, %d", e.stock(a.ID), e.stock(b.ID))
	}
	if cart, _ := e.svc.GetCart(e.ctx, u); len(cart.Items) != 2 {
		t.Fatalf("корзина не должна меняться: %+v", cart)
	}
	if mine, _ := e.svc.ListMyOrders(e.ctx, u); len(mine) != 0 {
		t.Fatalf("заказ не должен создаться: %+v", mine)
	}
	if len(e.gw.created) != 0 || len(e.db.events) != 0 {
		t.Fatal("не должно быть ни платежа, ни события")
	}
}

// failingEvents — публикация события падает: оформление должно откатиться целиком.
type failingEvents struct{}

func (failingEvents) Publish(context.Context, string, any) error {
	return errors.New("outbox недоступен")
}

func TestCheckoutRollbackOnPublishError(t *testing.T) {
	e := newEnv(t)
	e.svc.events = failingEvents{}
	u := newUserID()
	p := e.newProduct(10000, 5)
	e.put(u, p.ID, 2)
	if _, err := e.svc.Checkout(e.ctx, u, validCheckout); err == nil {
		t.Fatal("ожидалась ошибка")
	}
	if e.stock(p.ID) != 5 || len(e.db.orders) != 0 || len(e.db.carts[u]) != 1 || len(e.gw.created) != 0 {
		t.Fatal("неудачное оформление должно откатиться целиком")
	}
}

func TestCheckoutRejected(t *testing.T) {
	e := newEnv(t)
	u := newUserID()

	_, err := e.svc.Checkout(e.ctx, u, validCheckout)
	if !errors.Is(err, shop.ErrCartEmpty) {
		t.Fatalf("пустая корзина: %v", err)
	}

	p := e.newProduct(10000, 5)
	e.put(u, p.ID, 1)
	_, err = e.svc.Checkout(e.ctx, u, shop.CheckoutInput{DeliveryAddress: "  ", Phone: validCheckout.Phone})
	wantKind(t, err, errs.Invalid)
	_, err = e.svc.Checkout(e.ctx, u, shop.CheckoutInput{DeliveryAddress: validCheckout.DeliveryAddress, Phone: "12-34"})
	wantKind(t, err, errs.Invalid)

	// Заблокированный пользователь не может оформить заказ.
	e.users.blocked[u] = true
	_, err = e.svc.Checkout(e.ctx, u, validCheckout)
	wantKind(t, err, errs.Forbidden)
	if e.stock(p.ID) != 5 || len(e.db.orders) != 0 || len(e.db.events) != 0 {
		t.Fatal("ничего не должно измениться")
	}
}

func TestCheckoutWithoutPayments(t *testing.T) {
	e := newEnv(t)
	u := newUserID()
	p := e.newProduct(10000, 5)

	// ЮKassa не настроена: заказ оформляется, ссылки на оплату нет, Pay — 503.
	e.svc.payments = nil
	e.put(u, p.ID, 1)
	o, err := e.svc.Checkout(e.ctx, u, validCheckout)
	if err != nil || o.PaymentURL != "" || o.Status != shop.StatusPending {
		t.Fatalf("заказ без оплаты: %+v, %v", o, err)
	}
	_, err = e.svc.Pay(e.ctx, o.ID, u)
	if !errors.Is(err, shop.ErrPaymentsDisabled) {
		t.Fatalf("оплата без ЮKassa: %v", err)
	}

	// То же, если адаптер ЮKassa сообщает «не настроена».
	e.svc.payments = e.gw
	e.gw.err = ErrGatewayDisabled
	if _, err := e.svc.Pay(e.ctx, o.ID, u); !errors.Is(err, shop.ErrPaymentsDisabled) {
		t.Fatalf("ненастроенный шлюз: %v", err)
	}

	// ЮKassa недоступна при оформлении — ссылку можно получить позже через Pay.
	e.gw.err = errors.New("таймаут")
	e.put(u, p.ID, 1)
	o, err = e.svc.Checkout(e.ctx, u, validCheckout)
	if err != nil || o.PaymentURL != "" {
		t.Fatalf("заказ при недоступной ЮKassa: %+v, %v", o, err)
	}
	_, err = e.svc.Pay(e.ctx, o.ID, u)
	if !errors.Is(err, shop.ErrPaymentFailed) {
		t.Fatalf("недоступная ЮKassa: %v", err)
	}
	e.gw.err = nil
	o, err = e.svc.Pay(e.ctx, o.ID, u)
	if err != nil || o.PaymentURL == "" || o.PaymentID == "" {
		t.Fatalf("повторная оплата: %+v, %v", o, err)
	}
	_, err = e.svc.Pay(e.ctx, o.ID, newUserID())
	wantKind(t, err, errs.NotFound)
}

// checkoutOne оформляет заказ на qty штук нового товара (остаток 10).
func (e *env) checkoutOne(qty int) (shop.Order, shop.Product, int64) {
	e.t.Helper()
	u := newUserID()
	p := e.newProduct(250000, 10)
	e.put(u, p.ID, qty)
	o, err := e.svc.Checkout(e.ctx, u, validCheckout)
	if err != nil {
		e.t.Fatal(err)
	}
	return o, p, u
}

func (e *env) status(id int64) shop.Status { return e.db.orders[id].Status }

func (e *env) notify(paymentID string) {
	e.t.Helper()
	if err := e.svc.HandlePaymentNotification(e.ctx, paymentID); err != nil {
		e.t.Fatal(err)
	}
}

func TestWebhookSucceeded(t *testing.T) {
	e := newEnv(t)
	o, _, u := e.checkoutOne(2)
	payID := o.PaymentID

	// Платёж ещё не оплачен — уведомление ничего не меняет.
	e.notify(payID)
	if e.status(o.ID) != shop.StatusPending {
		t.Fatal("неоплаченный заказ не должен становиться оплаченным")
	}

	e.gw.set(payID, PaymentSucceeded, "")
	for range 3 { // ЮKassa может прислать уведомление несколько раз
		e.notify(payID)
	}
	if e.status(o.ID) != shop.StatusPaid {
		t.Fatal("заказ должен стать оплаченным")
	}
	paid := e.eventsOf(EventOrderPaid)
	if len(paid) != 1 || paid[0] != (OrderPaid{OrderID: o.ID, UserID: u, TotalKopecks: o.TotalKopecks}) {
		t.Fatalf("ожидалось ровно одно событие order.paid: %+v", paid)
	}

	// Неизвестный платёж — не ошибка, чтобы ЮKassa перестала его присылать.
	e.notify("no-such-payment")
	// Временный сбой ЮKassa — ошибка, чтобы ЮKassa повторила уведомление.
	e.gw.err = errors.New("ЮKassa недоступна")
	if err := e.svc.HandlePaymentNotification(e.ctx, payID); err == nil || errs.Is(err, errs.Invalid) {
		t.Fatalf("временный сбой: %v", err)
	}
	// Без шлюза уведомления игнорируются.
	e.svc.payments = nil
	e.notify(payID)
}

func TestWebhookAmountMismatch(t *testing.T) {
	e := newEnv(t)
	o, _, _ := e.checkoutOne(1)
	e.gw.set(o.PaymentID, PaymentSucceeded, "1.00")
	e.notify(o.PaymentID)
	if e.status(o.ID) != shop.StatusPending || len(e.eventsOf(EventOrderPaid)) != 0 {
		t.Fatal("платёж с другой суммой не должен оплачивать заказ")
	}
}

func TestWebhookCanceled(t *testing.T) {
	e := newEnv(t)
	o, p, u := e.checkoutOne(4)
	if e.stock(p.ID) != 6 {
		t.Fatalf("остаток после заказа: %d", e.stock(p.ID))
	}
	e.gw.set(o.PaymentID, PaymentCanceled, "")
	for range 3 {
		e.notify(o.PaymentID)
	}
	if e.status(o.ID) != shop.StatusCancelled {
		t.Fatal("заказ должен быть отменён")
	}
	if e.stock(p.ID) != 10 {
		t.Fatalf("товар должен вернуться на склад ровно один раз: остаток %d", e.stock(p.ID))
	}
	cancelled := e.eventsOf(EventOrderCancelled)
	if len(cancelled) != 1 || cancelled[0] != (OrderCancelled{OrderID: o.ID, UserID: u, TotalKopecks: o.TotalKopecks}) {
		t.Fatalf("события order.cancelled: %+v", cancelled)
	}
}

func TestAdminStatusTransitions(t *testing.T) {
	e := newEnv(t)

	// pending → paid вручную нельзя: оплату подтверждает только ЮKassa.
	o, p, _ := e.checkoutOne(3)
	_, err := e.svc.SetStatus(e.ctx, o.ID, "paid")
	wantKind(t, err, errs.Conflict)
	_, err = e.svc.SetStatus(e.ctx, o.ID, "shipped")
	wantKind(t, err, errs.Conflict)
	_, err = e.svc.SetStatus(e.ctx, o.ID, "lost")
	wantKind(t, err, errs.Invalid)
	_, err = e.svc.SetStatus(e.ctx, -1, "cancelled")
	wantKind(t, err, errs.NotFound)

	// pending → cancelled возвращает товар на склад.
	got, err := e.svc.SetStatus(e.ctx, o.ID, "cancelled")
	if err != nil || got.Status != shop.StatusCancelled || e.stock(p.ID) != 10 {
		t.Fatalf("отмена: %+v, %v, остаток %d", got, err, e.stock(p.ID))
	}
	_, err = e.svc.SetStatus(e.ctx, o.ID, "cancelled")
	wantKind(t, err, errs.Conflict)
	if e.stock(p.ID) != 10 || len(e.eventsOf(EventOrderCancelled)) != 1 {
		t.Fatal("повторная отмена не должна возвращать товар и публиковать событие")
	}

	// Оплата отменённого заказа не делает его оплаченным.
	e.gw.set(o.PaymentID, PaymentSucceeded, "")
	e.notify(o.PaymentID)
	if e.status(o.ID) != shop.StatusCancelled {
		t.Fatalf("оплата отменённого заказа: статус %s", e.status(o.ID))
	}

	// paid → shipped → completed; из completed никуда.
	o2, _, u2 := e.checkoutOne(1)
	e.gw.set(o2.PaymentID, PaymentSucceeded, "")
	e.notify(o2.PaymentID)
	for _, st := range []shop.Status{shop.StatusShipped, shop.StatusCompleted} {
		if got, err := e.svc.SetStatus(e.ctx, o2.ID, string(st)); err != nil || got.Status != st {
			t.Fatalf("переход в %s: %+v, %v", st, got, err)
		}
	}
	shipped := e.eventsOf(EventOrderShipped)
	if len(shipped) != 1 || shipped[0] != (OrderShipped{OrderID: o2.ID, UserID: u2, DeliveryAddress: validCheckout.DeliveryAddress}) {
		t.Fatalf("событие order.shipped: %+v", shipped)
	}
	_, err = e.svc.SetStatus(e.ctx, o2.ID, "cancelled")
	wantKind(t, err, errs.Conflict)

	// paid → cancelled тоже возвращает товар, в событии wasPaid.
	o3, p3, u3 := e.checkoutOne(2)
	e.gw.set(o3.PaymentID, PaymentSucceeded, "")
	e.notify(o3.PaymentID)
	if _, err := e.svc.SetStatus(e.ctx, o3.ID, "cancelled"); err != nil || e.stock(p3.ID) != 10 {
		t.Fatalf("отмена оплаченного: %v, остаток %d", err, e.stock(p3.ID))
	}
	cancelled := e.eventsOf(EventOrderCancelled)
	if last := cancelled[len(cancelled)-1]; last != (OrderCancelled{OrderID: o3.ID, UserID: u3, TotalKopecks: o3.TotalKopecks, WasPaid: true}) {
		t.Fatalf("событие отмены оплаченного: %+v", last)
	}

	// Список для администратора с фильтром по статусу.
	list, err := e.svc.ListOrders(e.ctx, "completed", 200, 0)
	if err != nil || len(list) != 1 || list[0].ID != o2.ID {
		t.Fatalf("фильтр completed: %+v, %v", list, err)
	}
	if all, _ := e.svc.ListOrders(e.ctx, "", 200, 0); len(all) != 3 {
		t.Fatalf("все заказы: %d", len(all))
	}
	_, err = e.svc.ListOrders(e.ctx, "lost", 50, 0)
	wantKind(t, err, errs.Invalid)
}

func TestCancelMine(t *testing.T) {
	e := newEnv(t)
	o, p, u := e.checkoutOne(3)

	_, err := e.svc.CancelMine(e.ctx, o.ID, newUserID())
	wantKind(t, err, errs.NotFound)

	got, err := e.svc.CancelMine(e.ctx, o.ID, u)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != shop.StatusCancelled || e.stock(p.ID) != 10 {
		t.Fatalf("после отмены: статус %s, остаток %d", got.Status, e.stock(p.ID))
	}
	_, err = e.svc.CancelMine(e.ctx, o.ID, u)
	if !errors.Is(err, shop.ErrCannotCancel) {
		t.Fatalf("повторная отмена: %v", err)
	}
	if len(e.eventsOf(EventOrderCancelled)) != 1 {
		t.Fatal("событие отмены должно быть одно")
	}
	// Отменённый заказ оплатить нельзя.
	if _, err := e.svc.Pay(e.ctx, o.ID, u); !errors.Is(err, shop.ErrOrderNotPending) {
		t.Fatalf("оплата отменённого: %v", err)
	}
}
