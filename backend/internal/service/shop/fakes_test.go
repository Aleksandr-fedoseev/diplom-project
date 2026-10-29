package shop

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"

	"heritage-backend/internal/domain/errs"
	"heritage-backend/internal/domain/shop"
)

// Фейки для тестов сервиса: всё хранится в памяти. Общее состояние (товары, корзины,
// заказы, события) лежит в memDB, а fakeTx при ошибке восстанавливает его снимок —
// так тесты проверяют, что неудачное оформление ничего не оставляет.

type event struct {
	Type    string
	Payload any
}

type memDB struct {
	products map[int64]shop.Product
	carts    map[int64][]shop.CartItem // по пользователю, в порядке добавления (данные товара берутся из products)
	orders   map[int64]shop.Order
	events   []event
	nextID   int64
	inTx     bool
}

func newMemDB() *memDB {
	return &memDB{products: map[int64]shop.Product{}, carts: map[int64][]shop.CartItem{}, orders: map[int64]shop.Order{}}
}

func (db *memDB) id() int64 { db.nextID++; return db.nextID }

// snapshot — глубокая копия состояния для отката.
func (db *memDB) snapshot() memDB {
	c := memDB{products: maps.Clone(db.products), carts: map[int64][]shop.CartItem{}, orders: map[int64]shop.Order{},
		events: slices.Clone(db.events), nextID: db.nextID}
	for k, v := range db.carts {
		c.carts[k] = slices.Clone(v)
	}
	for k, o := range db.orders {
		o.Items = slices.Clone(o.Items)
		c.orders[k] = o
	}
	return c
}

type fakeTx struct{ db *memDB }

func (t fakeTx) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if t.db.inTx { // вложенная транзакция — часть внешней
		return fn(ctx)
	}
	saved := t.db.snapshot()
	t.db.inTx = true
	err := fn(ctx)
	t.db.inTx = false
	if err != nil {
		*t.db = saved
	}
	return err
}

// Товары.

type fakeProducts struct{ db *memDB }

func (f fakeProducts) List(_ context.Context, category string, includeInactive bool) ([]shop.Product, error) {
	var out []shop.Product
	for _, p := range f.db.products {
		if (p.IsActive || includeInactive) && (category == "" || p.Category == category) {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f fakeProducts) BySlug(_ context.Context, slug string) (shop.Product, error) {
	for _, p := range f.db.products {
		if p.Slug == slug {
			return p, nil
		}
	}
	return shop.Product{}, shop.ErrProductNotFound
}

func (f fakeProducts) ByID(_ context.Context, id int64) (shop.Product, error) {
	p, ok := f.db.products[id]
	if !ok {
		return shop.Product{}, shop.ErrProductNotFound
	}
	return p, nil
}

func (f fakeProducts) slugTaken(slug string, exceptID int64) bool {
	for _, p := range f.db.products {
		if p.Slug == slug && p.ID != exceptID {
			return true
		}
	}
	return false
}

func (f fakeProducts) Create(_ context.Context, p shop.Product) (shop.Product, error) {
	if f.slugTaken(p.Slug, 0) {
		return shop.Product{}, shop.ErrSlugTaken
	}
	p.ID = f.db.id()
	f.db.products[p.ID] = p
	return p, nil
}

func (f fakeProducts) Update(_ context.Context, p shop.Product) (shop.Product, error) {
	if _, ok := f.db.products[p.ID]; !ok {
		return shop.Product{}, shop.ErrProductNotFound
	}
	if f.slugTaken(p.Slug, p.ID) {
		return shop.Product{}, shop.ErrSlugTaken
	}
	f.db.products[p.ID] = p
	return p, nil
}

func (f fakeProducts) Deactivate(_ context.Context, id int64) error {
	p, ok := f.db.products[id]
	if !ok {
		return shop.ErrProductNotFound
	}
	p.IsActive = false
	f.db.products[id] = p
	return nil
}

func (f fakeProducts) DecrementStock(_ context.Context, id int64, qty int) error {
	p := f.db.products[id]
	if p.Stock < qty {
		return errors.New("stock >= 0 нарушено")
	}
	p.Stock -= qty
	f.db.products[id] = p
	return nil
}

func (f fakeProducts) ReturnOrderStock(_ context.Context, orderID int64) error {
	for _, it := range f.db.orders[orderID].Items {
		p := f.db.products[it.ProductID]
		p.Stock += it.Quantity
		f.db.products[it.ProductID] = p
	}
	return nil
}

// Корзины.

type fakeCarts struct{ db *memDB }

func (f fakeCarts) Items(_ context.Context, userID int64) ([]shop.CartItem, error) {
	out := []shop.CartItem{}
	for _, it := range f.db.carts[userID] {
		p := f.db.products[it.ProductID]
		out = append(out, shop.CartItem{ProductID: p.ID, Slug: p.Slug, Title: p.Title, ImageURL: p.ImageURL,
			PriceKopecks: p.PriceKopecks, Quantity: it.Quantity, Stock: p.Stock, IsActive: p.IsActive})
	}
	return out, nil
}

func (f fakeCarts) Set(_ context.Context, userID, productID int64, qty int) error {
	items := f.db.carts[userID]
	for i := range items {
		if items[i].ProductID == productID {
			items[i].Quantity = qty
			return nil
		}
	}
	f.db.carts[userID] = append(items, shop.CartItem{ProductID: productID, Quantity: qty})
	return nil
}

func (f fakeCarts) Remove(_ context.Context, userID, productID int64) error {
	f.db.carts[userID] = slices.DeleteFunc(f.db.carts[userID], func(it shop.CartItem) bool { return it.ProductID == productID })
	return nil
}

func (f fakeCarts) Clear(_ context.Context, userID int64) error {
	delete(f.db.carts, userID)
	return nil
}

func (f fakeCarts) RemoveProduct(_ context.Context, productID int64) error {
	for u := range f.db.carts {
		_ = f.Remove(context.Background(), u, productID)
	}
	return nil
}

func (f fakeCarts) LockForCheckout(ctx context.Context, userID int64) ([]shop.CheckoutLine, error) {
	if !f.db.inTx {
		return nil, errors.New("LockForCheckout вне транзакции")
	}
	items, _ := f.Items(ctx, userID)
	sort.Slice(items, func(i, j int) bool { return items[i].ProductID < items[j].ProductID })
	lines := make([]shop.CheckoutLine, len(items))
	for i, it := range items {
		lines[i] = shop.CheckoutLine{ProductID: it.ProductID, Title: it.Title, PriceKopecks: it.PriceKopecks,
			Quantity: it.Quantity, Stock: it.Stock, IsActive: it.IsActive}
	}
	return lines, nil
}

// Заказы.

type fakeOrders struct{ db *memDB }

func (f fakeOrders) Create(_ context.Context, o shop.Order) (int64, error) {
	o.ID = f.db.id()
	o.Items = slices.Clone(o.Items)
	f.db.orders[o.ID] = o
	return o.ID, nil
}

func (f fakeOrders) ByID(_ context.Context, id int64) (shop.Order, error) {
	o, ok := f.db.orders[id]
	if !ok {
		return shop.Order{}, shop.ErrOrderNotFound
	}
	o.Items = slices.Clone(o.Items)
	return o, nil
}

func (f fakeOrders) ByPaymentID(ctx context.Context, paymentID string) (shop.Order, error) {
	for id, o := range f.db.orders {
		if o.PaymentID == paymentID && paymentID != "" {
			return f.ByID(ctx, id)
		}
	}
	return shop.Order{}, shop.ErrOrderNotFound
}

func (f fakeOrders) list(keep func(shop.Order) bool) []shop.Order {
	out := []shop.Order{}
	for _, o := range f.db.orders {
		if keep(o) {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

func (f fakeOrders) ListByUser(_ context.Context, userID int64) ([]shop.Order, error) {
	return f.list(func(o shop.Order) bool { return o.OwnedBy(userID) }), nil
}

func (f fakeOrders) List(_ context.Context, status shop.Status, limit, offset int) ([]shop.Order, error) {
	all := f.list(func(o shop.Order) bool { return status == "" || o.Status == status })
	if offset > len(all) {
		return []shop.Order{}, nil
	}
	return all[offset:min(len(all), offset+limit)], nil
}

func (f fakeOrders) SetPayment(_ context.Context, id int64, paymentID, url string) error {
	o := f.db.orders[id]
	o.PaymentID, o.PaymentURL = paymentID, url
	f.db.orders[id] = o
	return nil
}

func (f fakeOrders) ChangeStatus(_ context.Context, id int64, from, to shop.Status) (bool, error) {
	o, ok := f.db.orders[id]
	if !ok || o.Status != from {
		return false, nil
	}
	o.Status = to
	f.db.orders[id] = o
	return true, nil
}

// События.

type fakeEvents struct{ db *memDB }

func (f fakeEvents) Publish(_ context.Context, eventType string, payload any) error {
	if !f.db.inTx {
		return errors.New("событие публикуется вне транзакции")
	}
	f.db.events = append(f.db.events, event{eventType, payload})
	return nil
}

// Пользователи: заблокированные — в множестве blocked.

type fakeUsers struct{ blocked map[int64]bool }

func (f fakeUsers) EnsureActive(_ context.Context, id int64) error {
	if f.blocked[id] {
		return errs.New(errs.Forbidden, "пользователь заблокирован")
	}
	return nil
}

// fakeGateway изображает ЮKassa: запоминает созданные платежи, статус меняет тест.

type createdPayment struct {
	OrderID, Amount                        int64
	Description, ReturnURL, IdempotenceKey string
}

type fakeGateway struct {
	payments map[string]Payment
	created  []createdPayment
	err      error // если задана — все вызовы возвращают её
}

func newFakeGateway() *fakeGateway { return &fakeGateway{payments: map[string]Payment{}} }

func (g *fakeGateway) CreatePayment(_ context.Context, orderID, amount int64, description, returnURL, key string) (string, string, error) {
	if g.err != nil {
		return "", "", g.err
	}
	g.created = append(g.created, createdPayment{orderID, amount, description, returnURL, key})
	id := fmt.Sprintf("pay-%d-%d", orderID, len(g.created))
	g.payments[id] = Payment{ID: id, Status: "pending", Amount: shop.FormatRubles(amount), OrderID: fmt.Sprint(orderID)}
	return id, fmt.Sprintf("https://yoomoney.test/checkout?order=%d", orderID), nil
}

func (g *fakeGateway) GetPayment(_ context.Context, id string) (Payment, error) {
	if g.err != nil {
		return Payment{}, g.err
	}
	p, ok := g.payments[id]
	if !ok {
		return Payment{}, ErrGatewayPaymentNotFound
	}
	return p, nil
}

func (g *fakeGateway) set(id, status, amount string) {
	p := g.payments[id]
	p.Status = status
	if amount != "" {
		p.Amount = amount
	}
	g.payments[id] = p
}
