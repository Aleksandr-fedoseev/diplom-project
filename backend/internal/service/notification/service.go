// Package notification — сценарии модуля уведомлений: событие → письмо получателю.
// Реализует notification.Service из домена.
package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"heritage-backend/internal/domain/errs"
	"heritage-backend/internal/domain/notification"
)

type Service struct {
	users     Users
	mailer    Mailer
	publicURL string // адрес сайта для ссылок в письмах, без «/» в конце
}

// Проверка при сборке: Service реализует интерфейс домена.
var _ notification.Service = (*Service)(nil)

func NewService(users Users, mailer Mailer, publicURL string) *Service {
	return &Service{users: users, mailer: mailer, publicURL: strings.TrimRight(publicURL, "/")}
}

// Handlers — обработчик на каждое событие, на которое подписан модуль.
func (s *Service) Handlers() map[string]func(ctx context.Context, payload []byte) error {
	return map[string]func(ctx context.Context, payload []byte) error{
		notification.EventSubmissionApproved: handle(s.submissionApproved),
		notification.EventSubmissionRejected: handle(s.submissionRejected),
		notification.EventOrderPlaced:        handle(s.orderPlaced),
		notification.EventOrderPaid:          handle(s.orderPaid),
		notification.EventOrderShipped:       handle(s.orderShipped),
		notification.EventOrderCancelled:     handle(s.orderCancelled),
	}
}

// handle превращает обработчик конкретного события в обработчик JSON:
// раскодирует payload в структуру события T и вызывает fn.
func handle[T any](fn func(ctx context.Context, ev T) error) func(ctx context.Context, payload []byte) error {
	return func(ctx context.Context, payload []byte) error {
		var ev T
		if err := json.Unmarshal(payload, &ev); err != nil {
			return fmt.Errorf("%w: %v", notification.ErrMalformedEvent, err)
		}
		return fn(ctx, ev)
	}
}

// greeting — начало каждого письма.
func greeting(name string) string { return fmt.Sprintf("Здравствуйте, %s!\n\n", name) }

func (s *Service) submissionApproved(ctx context.Context, ev notification.SubmissionApproved) error {
	if ev.SubmissionID == 0 || ev.AuthorID == 0 {
		return fmt.Errorf("%w: нет submissionId или authorId", notification.ErrMalformedEvent)
	}
	link := s.publicURL + "/objects/" + ev.ObjectSlug
	return s.send(ctx, ev.AuthorID, func(name string) (string, string) {
		return "Ваша заявка одобрена", greeting(name) +
			fmt.Sprintf("Ваша заявка «%s» одобрена модератором. "+
				"Спасибо за вклад в сохранение наследия Астрахани!\n\nПосмотреть объект: %s\n", ev.Title, link)
	})
}

func (s *Service) submissionRejected(ctx context.Context, ev notification.SubmissionRejected) error {
	if ev.SubmissionID == 0 || ev.AuthorID == 0 {
		return fmt.Errorf("%w: нет submissionId или authorId", notification.ErrMalformedEvent)
	}
	return s.send(ctx, ev.AuthorID, func(name string) (string, string) {
		body := greeting(name) + fmt.Sprintf("Ваша заявка «%s» отклонена модератором.\n\nПричина: %s\n", ev.Title, ev.Reason)
		if ev.AuthorBlocked {
			body += "\nКроме того, ваша учётная запись заблокирована: подавать новые заявки нельзя.\n"
		} else {
			body += fmt.Sprintf("\nВы можете исправить сведения и подать заявку снова: %s/contribute\n", s.publicURL)
		}
		return "Ваша заявка отклонена", body
	})
}

// В событии order.placed нет состава заказа и адреса (см. контракт), поэтому письмо короче,
// чем в старом модуле shop: номер, сумма и что делать дальше.
func (s *Service) orderPlaced(ctx context.Context, ev notification.OrderPlaced) error {
	if ev.OrderID == 0 {
		return fmt.Errorf("%w: нет orderId", notification.ErrMalformedEvent)
	}
	return s.send(ctx, ev.UserID, func(name string) (string, string) {
		return fmt.Sprintf("Заказ №%d оформлен", ev.OrderID), greeting(name) +
			fmt.Sprintf("Ваш заказ №%d оформлен.\n\nИтого: %s ₽\n\nЗаказ будет передан в доставку после оплаты.\n",
				ev.OrderID, notification.FormatRubles(ev.TotalKopecks))
	})
}

func (s *Service) orderPaid(ctx context.Context, ev notification.OrderPaid) error {
	if ev.OrderID == 0 {
		return fmt.Errorf("%w: нет orderId", notification.ErrMalformedEvent)
	}
	return s.send(ctx, ev.UserID, func(name string) (string, string) {
		return fmt.Sprintf("Оплата заказа №%d получена", ev.OrderID), greeting(name) +
			fmt.Sprintf("Мы получили оплату заказа №%d на сумму %s ₽. Сообщим, когда заказ будет отправлен.\n",
				ev.OrderID, notification.FormatRubles(ev.TotalKopecks))
	})
}

func (s *Service) orderShipped(ctx context.Context, ev notification.OrderShipped) error {
	if ev.OrderID == 0 {
		return fmt.Errorf("%w: нет orderId", notification.ErrMalformedEvent)
	}
	return s.send(ctx, ev.UserID, func(name string) (string, string) {
		return fmt.Sprintf("Заказ №%d отправлен", ev.OrderID), greeting(name) +
			fmt.Sprintf("Заказ №%d передан в доставку по адресу: %s.\n", ev.OrderID, ev.DeliveryAddress)
	})
}

func (s *Service) orderCancelled(ctx context.Context, ev notification.OrderCancelled) error {
	if ev.OrderID == 0 {
		return fmt.Errorf("%w: нет orderId", notification.ErrMalformedEvent)
	}
	return s.send(ctx, ev.UserID, func(name string) (string, string) {
		text := fmt.Sprintf("Заказ №%d отменён.", ev.OrderID)
		if ev.WasPaid {
			text += " Деньги будут возвращены на карту, с которой производилась оплата."
		}
		return fmt.Sprintf("Заказ №%d отменён", ev.OrderID), greeting(name) + text + "\n"
	})
}

// send находит получателя, собирает письмо функцией build(имя) → (тема, текст) и отправляет.
// Получателя нет (userID = 0 или пользователь удалён) — письмо пропускается без ошибки:
// повтор тут не поможет. Другие ошибки возвращаются — outbox повторит событие позже.
func (s *Service) send(ctx context.Context, userID int64, build func(name string) (subject, body string)) error {
	if userID == 0 {
		slog.Info("письмо не отправлено: у события нет получателя")
		return nil
	}
	email, name, err := s.users.Contact(ctx, userID)
	if errs.Is(err, errs.NotFound) {
		slog.Info("письмо не отправлено: получатель не найден", "userId", userID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("поиск получателя %d: %w", userID, err)
	}
	subject, body := build(name)
	msg := notification.Message{To: email, Subject: subject, Body: body}
	if err := msg.Validate(); errors.Is(err, notification.ErrNoRecipientEmail) {
		slog.Info("письмо не отправлено: у получателя нет адреса", "userId", userID)
		return nil
	} else if err != nil {
		return err
	}
	if err := s.mailer.Send(ctx, msg.To, msg.Subject, msg.Body); err != nil {
		return fmt.Errorf("отправка письма: %w", err)
	}
	return nil
}
