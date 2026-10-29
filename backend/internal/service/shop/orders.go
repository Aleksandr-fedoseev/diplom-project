package shop

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	"heritage-backend/internal/domain/shop"
)

// Оформление заказа и оплата.

// Checkout оформляет заказ из корзины и создаёт платёж.
//
// Всё, что касается склада, делается в одной транзакции: строки товаров блокируются,
// наличие проверяется, заказ с позициями записывается, остатки уменьшаются, корзина
// очищается и публикуется событие order.placed. Ошибка на любом шаге откатывает всё.
//
// Платёж создаётся после фиксации транзакции: держать блокировки во время запроса
// к внешнему сервису нельзя. Если ЮKassa не ответила, заказ всё равно сохранён
// (paymentUrl пустой), а ссылку на оплату можно получить позже через Pay.
func (s *Service) Checkout(ctx context.Context, userID int64, in shop.CheckoutInput) (shop.Order, error) {
	if err := in.Normalize(); err != nil {
		return shop.Order{}, err
	}
	if err := s.users.EnsureActive(ctx, userID); err != nil {
		return shop.Order{}, err
	}

	var order shop.Order
	err := s.tx.Do(ctx, func(ctx context.Context) error {
		lines, err := s.carts.LockForCheckout(ctx, userID)
		if err != nil {
			return err
		}
		order, err = shop.NewOrder(userID, in, lines)
		if err != nil {
			return err
		}
		if order.ID, err = s.orders.Create(ctx, order); err != nil {
			return err
		}
		for _, it := range order.Items {
			if err := s.products.DecrementStock(ctx, it.ProductID, it.Quantity); err != nil {
				return err
			}
		}
		if err := s.carts.Clear(ctx, userID); err != nil {
			return err
		}
		items := make([]OrderPlacedItem, len(order.Items))
		for i, it := range order.Items {
			items[i] = OrderPlacedItem{Title: it.Title, Quantity: it.Quantity, PriceKopecks: it.PriceKopecks}
		}
		return s.events.Publish(ctx, EventOrderPlaced, OrderPlaced{
			OrderID: order.ID, UserID: userID, TotalKopecks: order.TotalKopecks,
			DeliveryAddress: order.DeliveryAddress, Items: items,
		})
	})
	if err != nil {
		return shop.Order{}, err
	}

	if err := s.createPayment(ctx, order.ID, order.TotalKopecks); err != nil {
		if errors.Is(err, ErrGatewayDisabled) {
			slog.Warn("ЮKassa не настроена: заказ оформлен без платежа", "order", order.ID)
		} else {
			slog.Error("платёж не создан, покупатель может повторить оплату позже", "order", order.ID, "err", err)
		}
	}
	return s.orders.ByID(ctx, order.ID)
}

// Pay возвращает заказ со ссылкой на оплату, создавая платёж, если его ещё нет
// (например, при оформлении ЮKassa была недоступна).
func (s *Service) Pay(ctx context.Context, orderID, userID int64) (shop.Order, error) {
	o, err := s.ownOrder(ctx, orderID, userID)
	if err != nil {
		return shop.Order{}, err
	}
	if o.Status != shop.StatusPending {
		return shop.Order{}, shop.ErrOrderNotPending
	}
	if o.PaymentURL == "" {
		err := s.createPayment(ctx, o.ID, o.TotalKopecks)
		if errors.Is(err, ErrGatewayDisabled) {
			return shop.Order{}, shop.ErrPaymentsDisabled
		}
		if err != nil {
			slog.Error("платёж не создан", "order", o.ID, "err", err)
			return shop.Order{}, shop.ErrPaymentFailed
		}
	}
	return s.orders.ByID(ctx, orderID)
}

// createPayment создаёт платёж и запоминает его id и страницу оплаты в заказе.
// Ключ идемпотентности привязан к заказу: повторный запрос вернёт тот же платёж.
func (s *Service) createPayment(ctx context.Context, orderID, totalKopecks int64) error {
	if s.payments == nil {
		return ErrGatewayDisabled
	}
	paymentID, confirmationURL, err := s.payments.CreatePayment(ctx, orderID, totalKopecks,
		fmt.Sprintf("Заказ №%d", orderID), s.orderReturnURL(orderID), fmt.Sprintf("order-%d", orderID))
	if err != nil {
		return err
	}
	return s.orders.SetPayment(ctx, orderID, paymentID, confirmationURL)
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

// HandlePaymentNotification обрабатывает уведомление ЮKassa о платеже paymentID.
//
// Уведомления ЮKassa не подписаны, поэтому телу запроса не доверяем: берём из него
// только id платежа и сами запрашиваем платёж у ЮKassa — действуем по статусу из ответа.
//
// Обработка идемпотентна: статус меняется условно (только из pending), и событие или
// возврат товаров на склад делаются, только если строка действительно изменилась.
// Ошибка возвращается только для временных сбоев — тогда ЮKassa повторит уведомление.
// Неизвестный платёж ошибкой не считается, чтобы ЮKassa перестала его присылать.
func (s *Service) HandlePaymentNotification(ctx context.Context, paymentID string) error {
	if s.payments == nil {
		slog.Warn("уведомление ЮKassa, но приём оплаты не настроен", "payment", paymentID)
		return nil
	}
	p, err := s.payments.GetPayment(ctx, paymentID)
	if errors.Is(err, ErrGatewayPaymentNotFound) || errors.Is(err, ErrGatewayDisabled) {
		slog.Warn("уведомление о неизвестном платеже", "payment", paymentID, "err", err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("запрос платежа %s: %w", paymentID, err)
	}

	o, err := s.orders.ByPaymentID(ctx, p.ID)
	if errors.Is(err, shop.ErrOrderNotFound) {
		slog.Warn("платёж не привязан ни к одному заказу", "payment", p.ID, "metadataOrder", p.OrderID)
		return nil
	}
	if err != nil {
		return err
	}

	switch p.Status {
	case PaymentSucceeded:
		if p.Amount != shop.FormatRubles(o.TotalKopecks) {
			slog.Error("сумма платежа не совпадает с суммой заказа", "order", o.ID, "payment", p.ID, "amount", p.Amount)
			return nil
		}
		changed, err := s.changeStatus(ctx, o, shop.StatusPending, shop.StatusPaid)
		if err != nil {
			return err
		}
		if !changed && o.Status == shop.StatusCancelled {
			// Покупатель оплатил уже отменённый заказ — деньги нужно вернуть вручную в кабинете ЮKassa.
			slog.Warn("оплачен отменённый заказ, нужен возврат", "order", o.ID, "payment", p.ID)
		}
	case PaymentCanceled:
		if _, err := s.changeStatus(ctx, o, shop.StatusPending, shop.StatusCancelled); err != nil {
			return err
		}
	default:
		// pending и waiting_for_capture: платёж ещё не завершён, ждём следующего уведомления.
	}
	return nil
}

// Заказы.

func (s *Service) GetOrder(ctx context.Context, id int64, viewer shop.Viewer) (shop.Order, error) {
	o, err := s.orders.ByID(ctx, id)
	if err != nil {
		return shop.Order{}, err
	}
	if !o.VisibleTo(viewer) {
		return shop.Order{}, shop.ErrOrderNotFound // чужой заказ — как будто его нет
	}
	return o, nil
}

// ownOrder — заказ покупателя; чужой отвечает «не найден».
func (s *Service) ownOrder(ctx context.Context, orderID, userID int64) (shop.Order, error) {
	return s.GetOrder(ctx, orderID, shop.Viewer{UserID: userID})
}

func (s *Service) ListMyOrders(ctx context.Context, userID int64) ([]shop.Order, error) {
	return s.orders.ListByUser(ctx, userID)
}

func (s *Service) ListOrders(ctx context.Context, status string, limit, offset int) ([]shop.Order, error) {
	var st shop.Status
	if status != "" {
		var err error
		if st, err = shop.ParseStatus(status); err != nil {
			return nil, err
		}
	}
	return s.orders.List(ctx, st, limit, offset)
}

// CancelMine — покупатель отменяет свой неоплаченный заказ, товары возвращаются на склад.
// Без этого брошенный заказ держал бы товар, пока платёж не истечёт в ЮKassa.
func (s *Service) CancelMine(ctx context.Context, orderID, userID int64) (shop.Order, error) {
	o, err := s.ownOrder(ctx, orderID, userID)
	if err != nil {
		return shop.Order{}, err
	}
	changed, err := s.changeStatus(ctx, o, shop.StatusPending, shop.StatusCancelled)
	if err != nil {
		return shop.Order{}, err
	}
	if !changed {
		return shop.Order{}, shop.ErrCannotCancel
	}
	return s.orders.ByID(ctx, orderID)
}

// SetStatus — смена статуса администратором (только разрешённые переходы, см. shop.CheckTransition).
func (s *Service) SetStatus(ctx context.Context, id int64, status string) (shop.Order, error) {
	to, err := shop.ParseStatus(status)
	if err != nil {
		return shop.Order{}, err
	}
	o, err := s.orders.ByID(ctx, id)
	if err != nil {
		return shop.Order{}, err
	}
	if err := shop.CheckTransition(o.Status, to); err != nil {
		return shop.Order{}, err
	}
	changed, err := s.changeStatus(ctx, o, o.Status, to)
	if err != nil {
		return shop.Order{}, err
	}
	if !changed {
		return shop.Order{}, shop.ErrStatusChangedRace
	}
	return s.orders.ByID(ctx, id)
}

// changeStatus в одной транзакции переводит заказ из from в to, при отмене возвращает
// товары на склад и публикует событие. Если заказ уже не в статусе from (повторное
// уведомление, параллельный запрос), ничего не делает и возвращает false — так
// событие и возврат на склад случаются ровно один раз. Из o берутся только данные для события.
func (s *Service) changeStatus(ctx context.Context, o shop.Order, from, to shop.Status) (bool, error) {
	changed := false
	err := s.tx.Do(ctx, func(ctx context.Context) error {
		var err error
		if changed, err = s.orders.ChangeStatus(ctx, o.ID, from, to); err != nil || !changed {
			return err
		}
		if to == shop.StatusCancelled {
			if err := s.products.ReturnOrderStock(ctx, o.ID); err != nil {
				return err
			}
		}
		return s.publishStatus(ctx, o, from, to)
	})
	if err != nil {
		return false, err
	}
	return changed, nil
}

// publishStatus публикует событие о новом статусе. Если покупатель удалён (UserID nil),
// уведомлять некого — событие не публикуется. О статусе «выполнен» событий нет.
func (s *Service) publishStatus(ctx context.Context, o shop.Order, from, to shop.Status) error {
	if o.UserID == nil {
		return nil
	}
	userID := *o.UserID
	switch to {
	case shop.StatusPaid:
		return s.events.Publish(ctx, EventOrderPaid, OrderPaid{OrderID: o.ID, UserID: userID, TotalKopecks: o.TotalKopecks})
	case shop.StatusShipped:
		return s.events.Publish(ctx, EventOrderShipped, OrderShipped{OrderID: o.ID, UserID: userID, DeliveryAddress: o.DeliveryAddress})
	case shop.StatusCancelled:
		return s.events.Publish(ctx, EventOrderCancelled, OrderCancelled{
			OrderID: o.ID, UserID: userID, TotalKopecks: o.TotalKopecks, WasPaid: from == shop.StatusPaid,
		})
	}
	return nil
}
