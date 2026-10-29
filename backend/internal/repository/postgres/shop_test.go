package postgres_test

import (
	"context"
	"errors"
	"testing"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/shop"
	"heritage-backend/internal/infrastructure/postgres"
	repo "heritage-backend/internal/repository/postgres"
	"heritage-backend/internal/testdb"
)

// Тест репозиториев магазина на настоящей БД: SQL, блокировка товаров при оформлении,
// условная смена статуса и перевод ошибок БД в ошибки домена.
func TestShopRepositories(t *testing.T) {
	pool := testdb.Connect(t)
	ctx := context.Background()
	products, carts, orders := repo.NewProducts(pool), repo.NewCarts(pool), repo.NewOrders(pool)
	tm := postgres.NewTxManager(pool)

	u, err := repo.NewUsers(pool).Create(ctx, auth.User{
		Email: testdb.Unique("shop") + "@example.ru", Name: "Покупатель", Role: auth.RoleUser, PasswordHash: "h",
	})
	if err != nil {
		t.Fatal(err)
	}

	newProduct := func(price int64, stock int) shop.Product {
		t.Helper()
		p, err := products.Create(ctx, shop.Product{
			Slug: testdb.Unique("repo-product"), Title: testdb.Unique("Товар"), Category: "repo-test",
			PriceKopecks: price, Stock: stock, IsActive: true,
		})
		if err != nil || p.ID == 0 || p.UpdatedAt.IsZero() {
			t.Fatalf("создание товара: %+v, %v", p, err)
		}
		return p
	}
	stock := func(id int64) int {
		t.Helper()
		p, err := products.ByID(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		return p.Stock
	}

	a := newProduct(450000, 5)
	b := newProduct(90000, 10)

	// Товары: дубликат slug, не найден, изменение.
	if _, err := products.Create(ctx, shop.Product{Slug: a.Slug, Title: "x", PriceKopecks: 1, IsActive: true}); !errors.Is(err, shop.ErrSlugTaken) {
		t.Fatalf("дубликат slug: %v", err)
	}
	if _, err := products.BySlug(ctx, "no-such-product"); !errors.Is(err, shop.ErrProductNotFound) {
		t.Fatalf("несуществующий товар: %v", err)
	}
	if _, err := products.Update(ctx, shop.Product{ID: -1, Slug: "x", Title: "x", PriceKopecks: 1}); !errors.Is(err, shop.ErrProductNotFound) {
		t.Fatalf("изменение несуществующего: %v", err)
	}
	if err := products.Deactivate(ctx, -1); !errors.Is(err, shop.ErrProductNotFound) {
		t.Fatalf("снятие несуществующего: %v", err)
	}

	// Корзина: b добавлен раньше a, но блокировка идёт в порядке id товара.
	if err := carts.Set(ctx, u.ID, b.ID, 3); err != nil {
		t.Fatal(err)
	}
	if err := carts.Set(ctx, u.ID, a.ID, 2); err != nil {
		t.Fatal(err)
	}
	if err := carts.Set(ctx, u.ID, -1, 1); !errors.Is(err, shop.ErrProductNotFound) {
		t.Fatalf("несуществующий товар в корзину: %v", err)
	}
	items, err := carts.Items(ctx, u.ID)
	if err != nil || len(items) != 2 || items[0].ProductID != b.ID || items[0].Quantity != 3 || items[1].Title != a.Title {
		t.Fatalf("корзина: %+v, %v", items, err)
	}

	// Оформление в транзакции: пока она открыта, строки товаров заблокированы.
	var orderID int64
	err = tm.Do(ctx, func(ctx context.Context) error {
		lines, err := carts.LockForCheckout(ctx, u.ID)
		if err != nil {
			return err
		}
		if len(lines) != 2 || lines[0].ProductID != a.ID || lines[1].ProductID != b.ID {
			t.Fatalf("строки оформления не в порядке id: %+v", lines)
		}
		// Другое соединение не может взять блокировку той же строки.
		if _, err := pool.Exec(context.Background(), `SELECT 1 FROM shop.products WHERE id = $1 FOR UPDATE NOWAIT`, a.ID); err == nil {
			t.Fatal("строка товара должна быть заблокирована до конца транзакции")
		}
		o, err := shop.NewOrder(u.ID, shop.CheckoutInput{DeliveryAddress: "адрес", Phone: "+79001234567"}, lines)
		if err != nil {
			return err
		}
		if orderID, err = orders.Create(ctx, o); err != nil {
			return err
		}
		for _, it := range o.Items {
			if err := products.DecrementStock(ctx, it.ProductID, it.Quantity); err != nil {
				return err
			}
		}
		return carts.Clear(ctx, u.ID)
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `SELECT 1 FROM shop.products WHERE id = $1 FOR UPDATE NOWAIT`, a.ID); err != nil {
		t.Fatalf("после фиксации блокировка должна сняться: %v", err)
	}
	if stock(a.ID) != 3 || stock(b.ID) != 7 {
		t.Fatalf("остатки: %d, %d", stock(a.ID), stock(b.ID))
	}
	if items, _ := carts.Items(ctx, u.ID); len(items) != 0 {
		t.Fatalf("корзина после заказа: %+v", items)
	}

	o, err := orders.ByID(ctx, orderID)
	if err != nil || o.Status != shop.StatusPending || o.TotalKopecks != 2*450000+3*90000 || len(o.Items) != 2 || !o.OwnedBy(u.ID) || o.PaymentID != "" {
		t.Fatalf("заказ: %+v, %v", o, err)
	}
	if _, err := orders.ByID(ctx, -1); !errors.Is(err, shop.ErrOrderNotFound) {
		t.Fatalf("несуществующий заказ: %v", err)
	}

	// Снимок цены: смена цены товара не меняет позиции заказа.
	a, _ = products.ByID(ctx, a.ID)
	a.PriceKopecks = 1
	if _, err := products.Update(ctx, a); err != nil {
		t.Fatal(err)
	}
	o, _ = orders.ByID(ctx, orderID)
	for _, it := range o.Items {
		if it.ProductID == a.ID && it.PriceKopecks != 450000 {
			t.Fatalf("снимок цены: %+v", it)
		}
	}

	// Платёж.
	payID := testdb.Unique("pay")
	if err := orders.SetPayment(ctx, orderID, payID, "https://pay.test"); err != nil {
		t.Fatal(err)
	}
	if got, err := orders.ByPaymentID(ctx, payID); err != nil || got.ID != orderID || got.PaymentURL != "https://pay.test" {
		t.Fatalf("по платежу: %+v, %v", got, err)
	}
	if _, err := orders.ByPaymentID(ctx, "no-such-payment"); !errors.Is(err, shop.ErrOrderNotFound) {
		t.Fatalf("неизвестный платёж: %v", err)
	}

	// Условная смена статуса: второй раз строка не меняется.
	if changed, err := orders.ChangeStatus(ctx, orderID, shop.StatusPending, shop.StatusCancelled); err != nil || !changed {
		t.Fatalf("первая смена статуса: %v, %v", changed, err)
	}
	if changed, err := orders.ChangeStatus(ctx, orderID, shop.StatusPending, shop.StatusCancelled); err != nil || changed {
		t.Fatalf("повторная смена статуса: %v, %v", changed, err)
	}
	if err := products.ReturnOrderStock(ctx, orderID); err != nil || stock(a.ID) != 5 || stock(b.ID) != 10 {
		t.Fatalf("возврат на склад: %v, %d, %d", err, stock(a.ID), stock(b.ID))
	}

	// Списки.
	mine, err := orders.ListByUser(ctx, u.ID)
	if err != nil || len(mine) != 1 || len(mine[0].Items) != 2 {
		t.Fatalf("мои заказы: %+v, %v", mine, err)
	}
	cancelled, err := orders.List(ctx, shop.StatusCancelled, 200, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, x := range cancelled {
		found = found || x.ID == orderID
		if x.Status != shop.StatusCancelled {
			t.Fatalf("в фильтре cancelled статус %s", x.Status)
		}
	}
	if !found {
		t.Fatal("отменённый заказ должен быть в списке")
	}

	// Снятие с продажи: витрина и корзины.
	if err := carts.Set(ctx, u.ID, b.ID, 1); err != nil {
		t.Fatal(err)
	}
	err = tm.Do(ctx, func(ctx context.Context) error {
		if err := products.Deactivate(ctx, b.ID); err != nil {
			return err
		}
		return carts.RemoveProduct(ctx, b.ID)
	})
	if err != nil {
		t.Fatal(err)
	}
	if items, _ := carts.Items(ctx, u.ID); len(items) != 0 {
		t.Fatalf("снятый товар остался в корзине: %+v", items)
	}
	visible, _ := products.List(ctx, "repo-test", false)
	all, _ := products.List(ctx, "repo-test", true)
	for _, p := range visible {
		if p.ID == b.ID {
			t.Fatal("снятый товар на витрине")
		}
	}
	if len(all) <= len(visible) {
		t.Fatal("администратор должен видеть и снятые товары")
	}
}
