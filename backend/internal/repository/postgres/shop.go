package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/domain/shop"
	"heritage-backend/internal/repository/postgres/sqlc"
)

// Репозитории магазина: товары, корзины, заказы (таблицы products, cart_items, orders, order_items).

// Товары.

type Products struct {
	pool *pgxpool.Pool
}

var _ shop.ProductRepository = (*Products)(nil)

func NewProducts(pool *pgxpool.Pool) *Products { return &Products{pool: pool} }

func toProduct(p sqlc.Product) shop.Product {
	return shop.Product{
		ID:           p.ID,
		Slug:         p.Slug,
		Title:        p.Title,
		Description:  p.Description,
		Category:     p.Category,
		PriceKopecks: p.PriceKopecks,
		Stock:        int(p.Stock),
		ImageURL:     p.ImageUrl,
		IsActive:     p.IsActive,
		UpdatedAt:    p.UpdatedAt,
	}
}

// oneProduct переводит результат запроса одной строки: нет строки — товар не найден,
// повтор slug — ErrSlugTaken.
func oneProduct(p sqlc.Product, err error) (shop.Product, error) {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return shop.Product{}, shop.ErrProductNotFound
	case pgCode(err) == codeUniqueViolation:
		return shop.Product{}, shop.ErrSlugTaken
	case err != nil:
		return shop.Product{}, err
	}
	return toProduct(p), nil
}

func (r *Products) List(ctx context.Context, category string, includeInactive bool) ([]shop.Product, error) {
	var cat *string
	if category != "" {
		cat = &category
	}
	rows, err := queries(ctx, r.pool).ListProducts(ctx, sqlc.ListProductsParams{IncludeInactive: includeInactive, Category: cat})
	if err != nil {
		return nil, err
	}
	list := make([]shop.Product, len(rows))
	for i, p := range rows {
		list[i] = toProduct(p)
	}
	return list, nil
}

func (r *Products) BySlug(ctx context.Context, slug string) (shop.Product, error) {
	return oneProduct(queries(ctx, r.pool).GetProductBySlug(ctx, slug))
}

func (r *Products) ByID(ctx context.Context, id int64) (shop.Product, error) {
	return oneProduct(queries(ctx, r.pool).GetProductByID(ctx, id))
}

func (r *Products) Create(ctx context.Context, p shop.Product) (shop.Product, error) {
	return oneProduct(queries(ctx, r.pool).CreateProduct(ctx, sqlc.CreateProductParams{
		Slug: p.Slug, Title: p.Title, Description: p.Description, Category: p.Category,
		PriceKopecks: p.PriceKopecks, Stock: int32(p.Stock), ImageUrl: p.ImageURL, IsActive: p.IsActive,
	}))
}

func (r *Products) Update(ctx context.Context, p shop.Product) (shop.Product, error) {
	return oneProduct(queries(ctx, r.pool).UpdateProduct(ctx, sqlc.UpdateProductParams{
		ID: p.ID, Slug: p.Slug, Title: p.Title, Description: p.Description, Category: p.Category,
		PriceKopecks: p.PriceKopecks, Stock: int32(p.Stock), ImageUrl: p.ImageURL, IsActive: p.IsActive,
	}))
}

func (r *Products) Deactivate(ctx context.Context, id int64) error {
	n, err := queries(ctx, r.pool).DeactivateProduct(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return shop.ErrProductNotFound
	}
	return nil
}

func (r *Products) DecrementStock(ctx context.Context, productID int64, quantity int) error {
	return queries(ctx, r.pool).DecrementProductStock(ctx, sqlc.DecrementProductStockParams{ID: productID, Quantity: int32(quantity)})
}

func (r *Products) ReturnOrderStock(ctx context.Context, orderID int64) error {
	return queries(ctx, r.pool).ReturnProductStockForOrder(ctx, orderID)
}

// Корзины.

type Carts struct {
	pool *pgxpool.Pool
}

var _ shop.CartRepository = (*Carts)(nil)

func NewCarts(pool *pgxpool.Pool) *Carts { return &Carts{pool: pool} }

func (r *Carts) Items(ctx context.Context, userID int64) ([]shop.CartItem, error) {
	rows, err := queries(ctx, r.pool).ListCartItems(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]shop.CartItem, len(rows))
	for i, c := range rows {
		items[i] = shop.CartItem{
			ProductID:    c.ProductID,
			Slug:         c.Slug,
			Title:        c.Title,
			ImageURL:     c.ImageUrl,
			PriceKopecks: c.PriceKopecks,
			Quantity:     int(c.Quantity),
			Stock:        int(c.Stock),
			IsActive:     c.IsActive,
		}
	}
	return items, nil
}

func (r *Carts) Set(ctx context.Context, userID, productID int64, quantity int) error {
	err := queries(ctx, r.pool).UpsertCartItem(ctx, sqlc.UpsertCartItemParams{UserID: userID, ProductID: productID, Quantity: int32(quantity)})
	if pgCode(err) == codeForeignKeyViolation {
		return shop.ErrProductNotFound // товар удалили между проверкой и записью
	}
	return err
}

func (r *Carts) Remove(ctx context.Context, userID, productID int64) error {
	return queries(ctx, r.pool).DeleteCartItem(ctx, sqlc.DeleteCartItemParams{UserID: userID, ProductID: productID})
}

func (r *Carts) Clear(ctx context.Context, userID int64) error {
	return queries(ctx, r.pool).ClearCart(ctx, userID)
}

func (r *Carts) RemoveProduct(ctx context.Context, productID int64) error {
	return queries(ctx, r.pool).DeleteCartItemsByProduct(ctx, productID)
}

func (r *Carts) LockForCheckout(ctx context.Context, userID int64) ([]shop.CheckoutLine, error) {
	rows, err := queries(ctx, r.pool).LockCartProducts(ctx, userID)
	if err != nil {
		return nil, err
	}
	lines := make([]shop.CheckoutLine, len(rows))
	for i, l := range rows {
		lines[i] = shop.CheckoutLine{
			ProductID:    l.ProductID,
			Title:        l.Title,
			PriceKopecks: l.PriceKopecks,
			Quantity:     int(l.Quantity),
			Stock:        int(l.Stock),
			IsActive:     l.IsActive,
		}
	}
	return lines, nil
}

// Заказы.

type Orders struct {
	pool *pgxpool.Pool
}

var _ shop.OrderRepository = (*Orders)(nil)

func NewOrders(pool *pgxpool.Pool) *Orders { return &Orders{pool: pool} }

func toOrder(o sqlc.Order) shop.Order {
	order := shop.Order{
		ID:              o.ID,
		UserID:          o.UserID,
		Status:          shop.Status(o.Status),
		TotalKopecks:    o.TotalKopecks,
		PaymentURL:      o.PaymentUrl,
		DeliveryAddress: o.DeliveryAddress,
		Phone:           o.Phone,
		Items:           []shop.OrderItem{},
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
	if o.PaymentID != nil {
		order.PaymentID = *o.PaymentID
	}
	return order
}

func (r *Orders) Create(ctx context.Context, o shop.Order) (int64, error) {
	q := queries(ctx, r.pool)
	id, err := q.CreateOrder(ctx, sqlc.CreateOrderParams{
		UserID: o.UserID, TotalKopecks: o.TotalKopecks, DeliveryAddress: o.DeliveryAddress, Phone: o.Phone,
	})
	if err != nil {
		return 0, err
	}
	for _, it := range o.Items {
		err := q.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
			OrderID: id, ProductID: it.ProductID, Title: it.Title, Quantity: int32(it.Quantity), PriceKopecks: it.PriceKopecks,
		})
		if err != nil {
			return 0, err
		}
	}
	return id, nil
}

// oneOrder — заказ одной строки вместе с позициями.
func (r *Orders) oneOrder(ctx context.Context, row sqlc.Order, err error) (shop.Order, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return shop.Order{}, shop.ErrOrderNotFound
	}
	if err != nil {
		return shop.Order{}, err
	}
	list, err := r.withItems(ctx, []sqlc.Order{row})
	if err != nil {
		return shop.Order{}, err
	}
	return list[0], nil
}

func (r *Orders) ByID(ctx context.Context, id int64) (shop.Order, error) {
	row, err := queries(ctx, r.pool).GetOrder(ctx, id)
	return r.oneOrder(ctx, row, err)
}

func (r *Orders) ByPaymentID(ctx context.Context, paymentID string) (shop.Order, error) {
	row, err := queries(ctx, r.pool).GetOrderByPaymentID(ctx, &paymentID)
	return r.oneOrder(ctx, row, err)
}

func (r *Orders) ListByUser(ctx context.Context, userID int64) ([]shop.Order, error) {
	rows, err := queries(ctx, r.pool).ListOrdersByUser(ctx, &userID)
	if err != nil {
		return nil, err
	}
	return r.withItems(ctx, rows)
}

func (r *Orders) List(ctx context.Context, status shop.Status, limit, offset int) ([]shop.Order, error) {
	var st *string
	if status != "" {
		s := string(status)
		st = &s
	}
	rows, err := queries(ctx, r.pool).ListOrders(ctx, sqlc.ListOrdersParams{Status: st, Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, err
	}
	return r.withItems(ctx, rows)
}

// withItems переводит строки заказов в сущности и заполняет позиции сразу для всех одним запросом.
func (r *Orders) withItems(ctx context.Context, rows []sqlc.Order) ([]shop.Order, error) {
	list := make([]shop.Order, len(rows))
	if len(rows) == 0 {
		return list, nil
	}
	ids := make([]int64, len(rows))
	for i, o := range rows {
		list[i] = toOrder(o)
		ids[i] = o.ID
	}
	items, err := queries(ctx, r.pool).ListOrderItems(ctx, ids)
	if err != nil {
		return nil, err
	}
	byOrder := make(map[int64][]shop.OrderItem, len(rows))
	for _, it := range items {
		byOrder[it.OrderID] = append(byOrder[it.OrderID], shop.OrderItem{
			ProductID: it.ProductID, Title: it.Title, Quantity: int(it.Quantity), PriceKopecks: it.PriceKopecks,
		})
	}
	for i := range list {
		if its := byOrder[list[i].ID]; its != nil {
			list[i].Items = its
		}
	}
	return list, nil
}

func (r *Orders) SetPayment(ctx context.Context, orderID int64, paymentID, paymentURL string) error {
	return queries(ctx, r.pool).SetOrderPayment(ctx, sqlc.SetOrderPaymentParams{ID: orderID, PaymentID: &paymentID, PaymentUrl: paymentURL})
}

func (r *Orders) ChangeStatus(ctx context.Context, orderID int64, from, to shop.Status) (bool, error) {
	n, err := queries(ctx, r.pool).ChangeOrderStatus(ctx, sqlc.ChangeOrderStatusParams{
		ID: orderID, OldStatus: string(from), NewStatus: string(to),
	})
	return n > 0, err
}
