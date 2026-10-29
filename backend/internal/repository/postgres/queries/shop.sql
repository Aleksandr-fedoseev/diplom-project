-- Магазин (модуль shop): товары, корзины, заказы.

-- Витрина показывает только товары в продаже; администратор видит и снятые (include_inactive).
-- name: ListProducts :many
SELECT * FROM products
WHERE (is_active OR sqlc.arg(include_inactive)::boolean)
  AND (sqlc.narg(category)::text IS NULL OR category = sqlc.narg(category)::text)
ORDER BY id;

-- name: GetProductBySlug :one
SELECT * FROM products WHERE slug = $1;

-- name: GetProductByID :one
SELECT * FROM products WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO products (slug, title, description, category, price_kopecks, stock, image_url, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateProduct :one
UPDATE products
SET slug = $2, title = $3, description = $4, category = $5, price_kopecks = $6,
    stock = $7, image_url = $8, is_active = $9, updated_at = now()
WHERE id = $1
RETURNING *;

-- Товар не удаляется физически: на него ссылаются позиции старых заказов.
-- name: DeactivateProduct :execrows
UPDATE products SET is_active = FALSE, updated_at = now() WHERE id = $1;

-- name: DecrementProductStock :exec
UPDATE products SET stock = stock - sqlc.arg(quantity)::int, updated_at = now() WHERE id = sqlc.arg(id);

-- Возврат товаров отменённого заказа на склад.
-- name: ReturnProductStockForOrder :exec
UPDATE products p
SET stock = p.stock + oi.quantity, updated_at = now()
FROM order_items oi
WHERE oi.order_id = $1 AND p.id = oi.product_id;

-- name: ListCartItems :many
SELECT c.product_id, c.quantity, p.slug, p.title, p.price_kopecks, p.stock, p.image_url, p.is_active
FROM cart_items c
JOIN products p ON p.id = c.product_id
WHERE c.user_id = $1
ORDER BY c.added_at, c.product_id;

-- name: UpsertCartItem :exec
INSERT INTO cart_items (user_id, product_id, quantity)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, product_id) DO UPDATE SET quantity = EXCLUDED.quantity;

-- name: DeleteCartItem :exec
DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2;

-- name: ClearCart :exec
DELETE FROM cart_items WHERE user_id = $1;

-- Снятый с продажи товар убираем из всех корзин.
-- name: DeleteCartItemsByProduct :exec
DELETE FROM cart_items WHERE product_id = $1;

-- Корзина вместе с товарами, строки товаров блокируются до конца транзакции:
-- два покупателя не смогут одновременно купить последний экземпляр.
-- Блокируем в порядке id, чтобы параллельные оформления не ждали друг друга по кругу (deadlock).
-- name: LockCartProducts :many
SELECT c.product_id, c.quantity, p.title, p.price_kopecks, p.stock, p.is_active
FROM cart_items c
JOIN products p ON p.id = c.product_id
WHERE c.user_id = $1
ORDER BY p.id
FOR UPDATE OF p;

-- name: CreateOrder :one
INSERT INTO orders (user_id, total_kopecks, delivery_address, phone)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: CreateOrderItem :exec
INSERT INTO order_items (order_id, product_id, title, quantity, price_kopecks)
VALUES ($1, $2, $3, $4, $5);

-- name: SetOrderPayment :exec
UPDATE orders SET payment_id = $2, payment_url = $3, updated_at = now() WHERE id = $1;

-- name: GetOrder :one
SELECT * FROM orders WHERE id = $1;

-- name: GetOrderByPaymentID :one
SELECT * FROM orders WHERE payment_id = $1;

-- name: ListOrdersByUser :many
SELECT * FROM orders WHERE user_id = $1 ORDER BY created_at DESC, id DESC;

-- name: ListOrders :many
SELECT * FROM orders
WHERE sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;

-- name: ListOrderItems :many
SELECT order_id, product_id, title, quantity, price_kopecks
FROM order_items
WHERE order_id = ANY(sqlc.arg(order_ids)::bigint[])
ORDER BY order_id, title;

-- Смена статуса срабатывает, только если заказ всё ещё в ожидаемом статусе.
-- Так повторное уведомление ЮKassa или параллельное действие администратора
-- ничего не меняют второй раз (например, не возвращают товары на склад дважды).
-- name: ChangeOrderStatus :execrows
UPDATE orders SET status = sqlc.arg(new_status), updated_at = now()
WHERE id = sqlc.arg(id) AND status = sqlc.arg(old_status);
