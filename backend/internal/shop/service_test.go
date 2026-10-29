package shop

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/auth"
	"heritage-backend/internal/httpx"
	"heritage-backend/internal/notifications"
	"heritage-backend/internal/testdb"
)

// fakeGateway изображает ЮKassa: запоминает созданные платежи, статус меняет тест.
type fakeGateway struct {
	mu       sync.Mutex
	payments map[string]Payment
	created  []PaymentRequest
	err      error // если задана — все вызовы возвращают её
}

func newFakeGateway() *fakeGateway { return &fakeGateway{payments: map[string]Payment{}} }

func (g *fakeGateway) CreatePayment(_ context.Context, req PaymentRequest) (Payment, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.err != nil {
		return Payment{}, g.err
	}
	g.created = append(g.created, req)
	p := Payment{
		ID:              testdb.Unique(fmt.Sprintf("pay-%d", req.OrderID)),
		Status:          PaymentPending,
		Amount:          formatAmount(req.AmountKopecks),
		ConfirmationURL: fmt.Sprintf("https://yoomoney.test/checkout?order=%d", req.OrderID),
		OrderID:         fmt.Sprint(req.OrderID),
	}
	g.payments[p.ID] = p
	return p, nil
}

func (g *fakeGateway) GetPayment(_ context.Context, id string) (Payment, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.err != nil {
		return Payment{}, g.err
	}
	p, ok := g.payments[id]
	if !ok {
		return Payment{}, ErrPaymentNotFound
	}
	p.ConfirmationURL = ""
	return p, nil
}

func (g *fakeGateway) setStatus(id, status string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	p := g.payments[id]
	p.Status = status
	g.payments[id] = p
}

// paymentIDByOrder находит id платежа, созданного для заказа.
func (g *fakeGateway) paymentIDByOrder(t *testing.T, orderID int64) string {
	t.Helper()
	g.mu.Lock()
	defer g.mu.Unlock()
	for id, p := range g.payments {
		if p.OrderID == fmt.Sprint(orderID) {
			return id
		}
	}
	t.Fatalf("платёж для заказа %d не создан", orderID)
	return ""
}

type nopSender struct{}

func (nopSender) Send(context.Context, notifications.Message) error { return nil }

type env struct {
	t     *testing.T
	ctx   context.Context
	svc   *Service
	users *auth.Service
	gw    *fakeGateway
}

func newEnv(t *testing.T) *env {
	db := testdb.Connect(t)
	users := auth.NewService(db)
	gw := newFakeGateway()
	svc := NewService(db, users, notifications.NewService(db, nopSender{}), gw, "https://site.test/shop/orders")
	return &env{t: t, ctx: context.Background(), svc: svc, users: users, gw: gw}
}

func (e *env) newUser() auth.User {
	e.t.Helper()
	u, err := e.users.Register(e.ctx, auth.RegisterInput{
		Email: testdb.Unique("buyer") + "@example.ru", Password: "password1", Name: "Покупатель",
	})
	if err != nil {
		e.t.Fatal(err)
	}
	return u
}

func (e *env) newProduct(price int64, stock int32) Product {
	e.t.Helper()
	p, err := e.svc.CreateProduct(e.ctx, ProductInput{
		Slug: testdb.Unique("test-product"), Title: "Тестовый товар " + testdb.Unique(""), Category: "test",
		PriceKopecks: price, Stock: stock,
	})
	if err != nil {
		e.t.Fatal(err)
	}
	return p
}

func (e *env) stock(id int64) int32 {
	e.t.Helper()
	p, err := e.svc.q.GetProductByID(e.ctx, id)
	if err != nil {
		e.t.Fatal(err)
	}
	return p.Stock
}

func (e *env) put(userID, productID int64, qty int32) Cart {
	e.t.Helper()
	cart, err := e.svc.SetCartItem(e.ctx, userID, productID, qty)
	if err != nil {
		e.t.Fatal(err)
	}
	return cart
}

// emails возвращает темы писем, поставленных в очередь на адрес.
func (e *env) emails(to string) []string {
	e.t.Helper()
	rows, err := e.svc.db.Query(e.ctx, `SELECT subject FROM email_jobs WHERE to_email = $1 ORDER BY id`, to)
	if err != nil {
		e.t.Fatal(err)
	}
	defer rows.Close()
	var subjects []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			e.t.Fatal(err)
		}
		subjects = append(subjects, s)
	}
	return subjects
}

var validCheckout = CheckoutInput{DeliveryAddress: "г. Астрахань, ул. Никольская, 1", Phone: "+7 (900) 123-45-67"}

func wantStatus(t *testing.T, err error, status int) {
	t.Helper()
	var he *httpx.Error
	if !errors.As(err, &he) || he.Status != status {
		t.Fatalf("ожидалась ошибка %d, получено: %v", status, err)
	}
}

func TestProducts(t *testing.T) {
	e := newEnv(t)
	p := e.newProduct(99900, 3)

	_, err := e.svc.CreateProduct(e.ctx, ProductInput{Slug: p.Slug, Title: "Дубликат", PriceKopecks: 100})
	wantStatus(t, err, 409)
	_, err = e.svc.CreateProduct(e.ctx, ProductInput{Slug: testdb.Unique("free"), Title: "Бесплатно", PriceKopecks: 0})
	wantStatus(t, err, 400)

	got, err := e.svc.GetProduct(e.ctx, p.Slug, false)
	if err != nil || got.PriceKopecks != 99900 || !got.InStock {
		t.Fatalf("карточка товара: %+v, %v", got, err)
	}
	list, err := e.svc.ListProducts(e.ctx, "test", false)
	if err != nil || !containsProduct(list, p.ID) {
		t.Fatalf("товар должен быть в категории test: %v", err)
	}

	// Снятый с продажи товар пропадает с витрины и из корзин, но остаётся в БД.
	u := e.newUser()
	e.put(u.ID, p.ID, 1)
	if err := e.svc.DeactivateProduct(e.ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.GetProduct(e.ctx, p.Slug, false)
	wantStatus(t, err, 404)
	if _, err := e.svc.GetProduct(e.ctx, p.Slug, true); err != nil {
		t.Fatalf("администратор должен видеть снятый товар: %v", err)
	}
	list, _ = e.svc.ListProducts(e.ctx, "test", false)
	if containsProduct(list, p.ID) {
		t.Fatal("снятый товар не должен быть на витрине")
	}
	if cart, _ := e.svc.GetCart(e.ctx, u.ID); len(cart.Items) != 0 {
		t.Fatalf("снятый товар должен пропасть из корзины: %+v", cart)
	}
	wantStatus(t, e.svc.DeactivateProduct(e.ctx, -1), 404)
}

func containsProduct(list []Product, id int64) bool {
	for _, p := range list {
		if p.ID == id {
			return true
		}
	}
	return false
}

func TestCart(t *testing.T) {
	e := newEnv(t)
	u := e.newUser()
	a := e.newProduct(450000, 5)
	b := e.newProduct(120000, 2)

	e.put(u.ID, a.ID, 2)
	cart := e.put(u.ID, b.ID, 1)
	if len(cart.Items) != 2 || cart.TotalKopecks != 2*450000+120000 {
		t.Fatalf("корзина: %+v", cart)
	}
	if cart.Items[0].ProductID != a.ID || cart.Items[0].LineTotalKopecks != 900000 || !cart.Items[0].Available {
		t.Fatalf("позиция корзины: %+v", cart.Items[0])
	}

	// Повторный PUT задаёт количество, а не прибавляет.
	cart = e.put(u.ID, a.ID, 5)
	if cart.Items[0].Quantity != 5 || cart.TotalKopecks != 5*450000+120000 {
		t.Fatalf("после изменения количества: %+v", cart)
	}

	_, err := e.svc.SetCartItem(e.ctx, u.ID, a.ID, 6)
	wantStatus(t, err, 409) // больше, чем на складе
	_, err = e.svc.SetCartItem(e.ctx, u.ID, a.ID, -1)
	wantStatus(t, err, 400)
	_, err = e.svc.SetCartItem(e.ctx, u.ID, -1, 1)
	wantStatus(t, err, 404)

	cart = e.put(u.ID, b.ID, 0) // 0 убирает товар
	if len(cart.Items) != 1 || cart.TotalKopecks != 5*450000 {
		t.Fatalf("после удаления количеством 0: %+v", cart)
	}
	cart, err = e.svc.RemoveCartItem(e.ctx, u.ID, a.ID)
	if err != nil || len(cart.Items) != 0 || cart.TotalKopecks != 0 {
		t.Fatalf("пустая корзина: %+v, %v", cart, err)
	}
}

func TestCheckout(t *testing.T) {
	e := newEnv(t)
	u := e.newUser()
	a := e.newProduct(450000, 5)
	b := e.newProduct(90000, 10)
	e.put(u.ID, a.ID, 2)
	e.put(u.ID, b.ID, 3)

	o, err := e.svc.Checkout(e.ctx, u.ID, validCheckout)
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != StatusPending || o.TotalKopecks != 2*450000+3*90000 || len(o.Items) != 2 ||
		o.DeliveryAddress != validCheckout.DeliveryAddress || o.Phone != validCheckout.Phone {
		t.Fatalf("заказ: %+v", o)
	}
	if o.PaymentURL != fmt.Sprintf("https://yoomoney.test/checkout?order=%d", o.ID) {
		t.Fatalf("ссылка на оплату: %q", o.PaymentURL)
	}
	req := e.gw.created[0]
	if req.AmountKopecks != o.TotalKopecks || req.ReturnURL != fmt.Sprintf("https://site.test/shop/orders?order=%d", o.ID) ||
		req.Description != fmt.Sprintf("Заказ №%d", o.ID) {
		t.Fatalf("запрос платежа: %+v", req)
	}
	if e.stock(a.ID) != 3 || e.stock(b.ID) != 7 {
		t.Fatalf("остатки после заказа: %d, %d", e.stock(a.ID), e.stock(b.ID))
	}
	if cart, _ := e.svc.GetCart(e.ctx, u.ID); len(cart.Items) != 0 {
		t.Fatalf("корзина после заказа должна быть пустой: %+v", cart)
	}

	// Позиции хранят цену и название на момент покупки: смена цены их не меняет.
	_, err = e.svc.UpdateProduct(e.ctx, a.ID, ProductInput{Slug: a.Slug, Title: "Новое название", PriceKopecks: 1, Stock: 3})
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.svc.GetOrder(e.ctx, o.ID, auth.Principal{UserID: u.ID, Role: auth.RoleUser})
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range got.Items {
		if it.ProductID == a.ID && (it.PriceKopecks != 450000 || it.Title != a.Title || it.LineTotalKopecks != 900000) {
			t.Fatalf("снимок позиции: %+v", it)
		}
	}

	// Чужой заказ не виден, администратору виден.
	other := e.newUser()
	_, err = e.svc.GetOrder(e.ctx, o.ID, auth.Principal{UserID: other.ID, Role: auth.RoleUser})
	wantStatus(t, err, 404)
	if _, err := e.svc.GetOrder(e.ctx, o.ID, auth.Principal{UserID: other.ID, Role: auth.RoleAdmin}); err != nil {
		t.Fatal(err)
	}

	mine, err := e.svc.ListMyOrders(e.ctx, u.ID)
	if err != nil || len(mine) != 1 || mine[0].ID != o.ID || len(mine[0].Items) != 2 {
		t.Fatalf("мои заказы: %+v, %v", mine, err)
	}
	if subj := e.emails(u.Email); len(subj) != 1 || !strings.Contains(subj[0], "оформлен") {
		t.Fatalf("письма: %v", subj)
	}

	// Pay для заказа с платежом не создаёт второй платёж.
	if _, err := e.svc.Pay(e.ctx, o.ID, u.ID); err != nil || len(e.gw.created) != 1 {
		t.Fatalf("повторная оплата: %v, платежей %d", err, len(e.gw.created))
	}
}

func TestCheckoutInsufficientStock(t *testing.T) {
	e := newEnv(t)
	u := e.newUser()
	a := e.newProduct(10000, 5)
	b := e.newProduct(20000, 2)
	e.put(u.ID, a.ID, 3)
	e.put(u.ID, b.ID, 2)

	// Пока товар лежал в корзине, остаток уменьшился (купили другие).
	if _, err := e.svc.UpdateProduct(e.ctx, b.ID, ProductInput{Slug: b.Slug, Title: b.Title, PriceKopecks: 20000, Stock: 1}); err != nil {
		t.Fatal(err)
	}
	cart, _ := e.svc.GetCart(e.ctx, u.ID)
	if cart.Items[1].Available {
		t.Fatalf("позиция без достаточного остатка должна быть недоступна: %+v", cart.Items[1])
	}

	_, err := e.svc.Checkout(e.ctx, u.ID, validCheckout)
	wantStatus(t, err, 409)

	// Ничего не изменилось: остатки, корзина, заказов нет.
	if e.stock(a.ID) != 5 || e.stock(b.ID) != 1 {
		t.Fatalf("остатки не должны меняться: %d, %d", e.stock(a.ID), e.stock(b.ID))
	}
	if cart, _ := e.svc.GetCart(e.ctx, u.ID); len(cart.Items) != 2 {
		t.Fatalf("корзина не должна меняться: %+v", cart)
	}
	if mine, _ := e.svc.ListMyOrders(e.ctx, u.ID); len(mine) != 0 {
		t.Fatalf("заказ не должен создаться: %+v", mine)
	}
	if len(e.gw.created) != 0 || len(e.emails(u.Email)) != 0 {
		t.Fatal("не должно быть ни платежа, ни письма")
	}
}

func TestCheckoutRejected(t *testing.T) {
	e := newEnv(t)
	u := e.newUser()

	_, err := e.svc.Checkout(e.ctx, u.ID, validCheckout)
	wantStatus(t, err, 400) // пустая корзина

	p := e.newProduct(10000, 5)
	e.put(u.ID, p.ID, 1)
	_, err = e.svc.Checkout(e.ctx, u.ID, CheckoutInput{DeliveryAddress: "  ", Phone: validCheckout.Phone})
	wantStatus(t, err, 400)
	_, err = e.svc.Checkout(e.ctx, u.ID, CheckoutInput{DeliveryAddress: validCheckout.DeliveryAddress, Phone: "12-34"})
	wantStatus(t, err, 400)
	_, err = e.svc.Checkout(e.ctx, u.ID, CheckoutInput{DeliveryAddress: validCheckout.DeliveryAddress, Phone: "+7 900 abc 45 67"})
	wantStatus(t, err, 400)

	// Заблокированный пользователь не может оформить заказ.
	admin := e.newUser()
	if _, err := e.users.SetBlocked(e.ctx, admin.ID, auth.RoleAdmin, u.ID, true); err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.Checkout(e.ctx, u.ID, validCheckout)
	wantStatus(t, err, 403)
	if e.stock(p.ID) != 5 {
		t.Fatal("остаток не должен меняться")
	}
}

func TestCheckoutWithoutPayments(t *testing.T) {
	e := newEnv(t)
	u := e.newUser()
	p := e.newProduct(10000, 5)

	// ЮKassa не настроена: заказ оформляется, ссылки на оплату нет.
	e.svc.payments = NewYooKassaClient("", "")
	e.put(u.ID, p.ID, 1)
	o, err := e.svc.Checkout(e.ctx, u.ID, validCheckout)
	if err != nil || o.PaymentURL != "" || o.Status != StatusPending {
		t.Fatalf("заказ без оплаты: %+v, %v", o, err)
	}
	_, err = e.svc.Pay(e.ctx, o.ID, u.ID)
	wantStatus(t, err, 503)

	// ЮKassa недоступна при оформлении — ссылку можно получить позже через Pay.
	e.svc.payments = e.gw
	e.gw.err = errors.New("таймаут")
	e.put(u.ID, p.ID, 1)
	o, err = e.svc.Checkout(e.ctx, u.ID, validCheckout)
	if err != nil || o.PaymentURL != "" {
		t.Fatalf("заказ при недоступной ЮKassa: %+v, %v", o, err)
	}
	_, err = e.svc.Pay(e.ctx, o.ID, u.ID)
	wantStatus(t, err, 502)
	e.gw.err = nil
	o, err = e.svc.Pay(e.ctx, o.ID, u.ID)
	if err != nil || o.PaymentURL == "" {
		t.Fatalf("повторная оплата: %+v, %v", o, err)
	}
	_, err = e.svc.Pay(e.ctx, o.ID, e.newUser().ID)
	wantStatus(t, err, 404)
}

// checkoutOne оформляет заказ на qty штук нового товара и возвращает заказ, товар и покупателя.
func (e *env) checkoutOne(qty int32) (Order, Product, auth.User) {
	e.t.Helper()
	u := e.newUser()
	p := e.newProduct(250000, 10)
	e.put(u.ID, p.ID, qty)
	o, err := e.svc.Checkout(e.ctx, u.ID, validCheckout)
	if err != nil {
		e.t.Fatal(err)
	}
	return o, p, u
}

func (e *env) orderStatus(id int64) string {
	e.t.Helper()
	o, err := e.svc.getOrder(e.ctx, id)
	if err != nil {
		e.t.Fatal(err)
	}
	return o.Status
}

func TestWebhookSucceeded(t *testing.T) {
	e := newEnv(t)
	o, _, u := e.checkoutOne(2)
	payID := e.gw.paymentIDByOrder(t, o.ID)

	// Платёж ещё не оплачен — уведомление ничего не меняет (тело уведомления не важно, важен ответ API).
	if err := e.svc.HandlePaymentNotification(e.ctx, payID); err != nil {
		t.Fatal(err)
	}
	if e.orderStatus(o.ID) != StatusPending {
		t.Fatal("неоплаченный заказ не должен становиться оплаченным")
	}

	e.gw.setStatus(payID, PaymentSucceeded)
	for range 3 { // ЮKassa может прислать уведомление несколько раз
		if err := e.svc.HandlePaymentNotification(e.ctx, payID); err != nil {
			t.Fatal(err)
		}
	}
	if e.orderStatus(o.ID) != StatusPaid {
		t.Fatal("заказ должен стать оплаченным")
	}
	subj := e.emails(u.Email)
	if len(subj) != 2 || !strings.Contains(subj[1], "Оплата") {
		t.Fatalf("ожидалось ровно одно письмо об оплате: %v", subj)
	}

	// Неизвестный платёж — не ошибка, чтобы ЮKassa перестала его присылать.
	if err := e.svc.HandlePaymentNotification(e.ctx, "no-such-payment"); err != nil {
		t.Fatal(err)
	}
}

func TestWebhookCanceled(t *testing.T) {
	e := newEnv(t)
	o, p, u := e.checkoutOne(4)
	if e.stock(p.ID) != 6 {
		t.Fatalf("остаток после заказа: %d", e.stock(p.ID))
	}
	payID := e.gw.paymentIDByOrder(t, o.ID)
	e.gw.setStatus(payID, PaymentCanceled)

	for range 3 {
		if err := e.svc.HandlePaymentNotification(e.ctx, payID); err != nil {
			t.Fatal(err)
		}
	}
	if e.orderStatus(o.ID) != StatusCancelled {
		t.Fatal("заказ должен быть отменён")
	}
	if e.stock(p.ID) != 10 {
		t.Fatalf("товар должен вернуться на склад ровно один раз: остаток %d", e.stock(p.ID))
	}
	if subj := e.emails(u.Email); len(subj) != 2 || !strings.Contains(subj[1], "отменён") {
		t.Fatalf("письма: %v", subj)
	}
}

func TestWebhookHandler(t *testing.T) {
	e := newEnv(t)
	o, _, _ := e.checkoutOne(1)
	payID := e.gw.paymentIDByOrder(t, o.ID)
	e.gw.setStatus(payID, PaymentSucceeded)

	r := chi.NewRouter()
	NewHandler(e.svc).Routes(r)
	post := func(body string) int {
		req := httptest.NewRequest(http.MethodPost, "/payments/yookassa/webhook", strings.NewReader(body))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec.Code
	}

	// В теле написано «canceled», но статус берётся у ЮKassa — заказ оплачен.
	notification := `{"type":"notification","event":"payment.canceled","object":{"id":"` + payID + `","status":"canceled"}}`
	if code := post(notification); code != 200 {
		t.Fatalf("уведомление: HTTP %d", code)
	}
	if e.orderStatus(o.ID) != StatusPaid {
		t.Fatal("статус должен браться из API ЮKassa, а не из тела уведомления")
	}
	if code := post(`{"type":"notification","event":"payment.succeeded","object":{"id":"unknown"}}`); code != 200 {
		t.Fatalf("неизвестный платёж: HTTP %d", code)
	}
	if code := post(`{}`); code != 400 {
		t.Fatalf("уведомление без id: HTTP %d", code)
	}
	e.gw.err = errors.New("ЮKassa недоступна")
	if code := post(notification); code != 500 {
		t.Fatalf("временный сбой должен давать 5xx, чтобы ЮKassa повторила: HTTP %d", code)
	}
}

func TestAdminStatusTransitions(t *testing.T) {
	e := newEnv(t)

	// pending → paid вручную нельзя: оплату подтверждает только ЮKassa.
	o, p, u := e.checkoutOne(3)
	_, err := e.svc.SetStatus(e.ctx, o.ID, StatusPaid)
	wantStatus(t, err, 409)
	_, err = e.svc.SetStatus(e.ctx, o.ID, StatusShipped)
	wantStatus(t, err, 409)
	_, err = e.svc.SetStatus(e.ctx, o.ID, "lost")
	wantStatus(t, err, 400)
	_, err = e.svc.SetStatus(e.ctx, -1, StatusCancelled)
	wantStatus(t, err, 404)

	// pending → cancelled возвращает товар на склад.
	got, err := e.svc.SetStatus(e.ctx, o.ID, StatusCancelled)
	if err != nil || got.Status != StatusCancelled || e.stock(p.ID) != 10 {
		t.Fatalf("отмена: %+v, %v, остаток %d", got, err, e.stock(p.ID))
	}
	_, err = e.svc.SetStatus(e.ctx, o.ID, StatusCancelled)
	wantStatus(t, err, 409)
	if e.stock(p.ID) != 10 {
		t.Fatal("повторная отмена не должна возвращать товар")
	}

	// Оплата отменённого заказа не делает его оплаченным.
	payID := e.gw.paymentIDByOrder(t, o.ID)
	e.gw.setStatus(payID, PaymentSucceeded)
	if err := e.svc.HandlePaymentNotification(e.ctx, payID); err != nil || e.orderStatus(o.ID) != StatusCancelled {
		t.Fatalf("оплата отменённого заказа: %v, статус %s", err, e.orderStatus(o.ID))
	}

	// paid → shipped → completed; из completed никуда.
	o2, _, _ := e.checkoutOne(1)
	pay2 := e.gw.paymentIDByOrder(t, o2.ID)
	e.gw.setStatus(pay2, PaymentSucceeded)
	if err := e.svc.HandlePaymentNotification(e.ctx, pay2); err != nil {
		t.Fatal(err)
	}
	for _, st := range []string{StatusShipped, StatusCompleted} {
		if got, err := e.svc.SetStatus(e.ctx, o2.ID, st); err != nil || got.Status != st {
			t.Fatalf("переход в %s: %+v, %v", st, got, err)
		}
	}
	_, err = e.svc.SetStatus(e.ctx, o2.ID, StatusCancelled)
	wantStatus(t, err, 409)

	// paid → cancelled тоже возвращает товар.
	o3, p3, _ := e.checkoutOne(2)
	pay3 := e.gw.paymentIDByOrder(t, o3.ID)
	e.gw.setStatus(pay3, PaymentSucceeded)
	if err := e.svc.HandlePaymentNotification(e.ctx, pay3); err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.SetStatus(e.ctx, o3.ID, StatusCancelled); err != nil || e.stock(p3.ID) != 10 {
		t.Fatalf("отмена оплаченного: %v, остаток %d", err, e.stock(p3.ID))
	}

	// Список для администратора с фильтром по статусу.
	list, err := e.svc.ListOrders(e.ctx, StatusCompleted, 200, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, x := range list {
		found = found || x.ID == o2.ID
		if x.Status != StatusCompleted {
			t.Fatalf("в фильтре completed заказ со статусом %s", x.Status)
		}
	}
	if !found {
		t.Fatal("выполненный заказ должен быть в списке")
	}
	_, err = e.svc.ListOrders(e.ctx, "lost", 50, 0)
	wantStatus(t, err, 400)

	if subj := e.emails(u.Email); len(subj) != 2 || !strings.Contains(subj[1], "отменён") {
		t.Fatalf("письма покупателю: %v", subj)
	}
}

func TestCancelMine(t *testing.T) {
	e := newEnv(t)
	o, p, u := e.checkoutOne(3)

	other := e.newUser()
	_, err := e.svc.CancelMine(e.ctx, o.ID, other.ID)
	wantStatus(t, err, 404)

	got, err := e.svc.CancelMine(e.ctx, o.ID, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusCancelled || e.stock(p.ID) != 10 {
		t.Fatalf("после отмены: статус %s, остаток %d", got.Status, e.stock(p.ID))
	}
	_, err = e.svc.CancelMine(e.ctx, o.ID, u.ID)
	wantStatus(t, err, 409)
}
