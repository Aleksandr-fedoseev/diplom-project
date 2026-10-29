package notification

import (
	"context"
	"errors"
	"strings"
	"testing"

	"heritage-backend/internal/domain/errs"
	"heritage-backend/internal/domain/notification"
)

// Тесты сервиса без БД и SMTP: пользователи и почта — фейки в памяти.

type contact struct{ email, name string }

type fakeUsers struct {
	byID map[int64]contact
	err  error // если задана — возвращается на любой запрос
}

func (f *fakeUsers) Contact(_ context.Context, id int64) (string, string, error) {
	if f.err != nil {
		return "", "", f.err
	}
	c, ok := f.byID[id]
	if !ok {
		return "", "", notification.ErrRecipientNotFound
	}
	return c.email, c.name, nil
}

type sent struct{ to, subject, body string }

type fakeMailer struct {
	sent []sent
	err  error
}

func (f *fakeMailer) Send(_ context.Context, to, subject, body string) error {
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, sent{to, subject, body})
	return nil
}

func newTestService() (*Service, *fakeUsers, *fakeMailer) {
	users := &fakeUsers{byID: map[int64]contact{
		7: {"anna@example.ru", "Анна"},
		8: {"", "Без почты"},
	}}
	mailer := &fakeMailer{}
	return NewService(users, mailer, "https://heritage.example/"), users, mailer
}

func TestEventsProduceEmails(t *testing.T) {
	cases := []struct {
		event   string
		payload string
		subject string
		body    []string // фрагменты текста письма
		absent  []string // чего в письме быть не должно
	}{
		{
			event:   notification.EventSubmissionApproved,
			payload: `{"submissionId":1,"authorId":7,"title":"Дом Тетюшинова","objectSlug":"dom-tetyushinova"}`,
			subject: "Ваша заявка одобрена",
			body: []string{"Здравствуйте, Анна!", "Ваша заявка «Дом Тетюшинова» одобрена модератором.",
				"Посмотреть объект: https://heritage.example/objects/dom-tetyushinova\n"},
		},
		{
			event:   notification.EventSubmissionRejected,
			payload: `{"submissionId":2,"authorId":7,"title":"Башня","reason":"нет источника","authorBlocked":false}`,
			subject: "Ваша заявка отклонена",
			body: []string{"Здравствуйте, Анна!", "Ваша заявка «Башня» отклонена модератором.", "Причина: нет источника",
				"подать заявку снова: https://heritage.example/contribute\n"},
			absent: []string{"заблокирована"},
		},
		{
			event:   notification.EventSubmissionRejected,
			payload: `{"submissionId":3,"authorId":7,"title":"Спам","reason":"реклама","authorBlocked":true}`,
			subject: "Ваша заявка отклонена",
			body:    []string{"Причина: реклама", "ваша учётная запись заблокирована: подавать новые заявки нельзя."},
			absent:  []string{"/contribute"},
		},
		{
			event:   notification.EventOrderPlaced,
			payload: `{"orderId":15,"userId":7,"totalKopecks":450000}`,
			subject: "Заказ №15 оформлен",
			body:    []string{"Здравствуйте, Анна!", "Ваш заказ №15 оформлен.", "Итого: 4500.00 ₽", "после оплаты"},
		},
		{
			event:   notification.EventOrderPaid,
			payload: `{"orderId":15,"userId":7,"totalKopecks":195005}`,
			subject: "Оплата заказа №15 получена",
			body:    []string{"Мы получили оплату заказа №15 на сумму 1950.05 ₽."},
		},
		{
			event:   notification.EventOrderShipped,
			payload: `{"orderId":15,"userId":7,"deliveryAddress":"Астрахань, ул. Советская, 1"}`,
			subject: "Заказ №15 отправлен",
			body:    []string{"Заказ №15 передан в доставку по адресу: Астрахань, ул. Советская, 1."},
		},
		{
			event:   notification.EventOrderCancelled,
			payload: `{"orderId":15,"userId":7,"totalKopecks":450000,"wasPaid":false}`,
			subject: "Заказ №15 отменён",
			body:    []string{"Заказ №15 отменён."},
			absent:  []string{"возвращены"},
		},
		{
			event:   notification.EventOrderCancelled,
			payload: `{"orderId":15,"userId":7,"totalKopecks":450000,"wasPaid":true}`,
			subject: "Заказ №15 отменён",
			body:    []string{"Заказ №15 отменён. Деньги будут возвращены на карту, с которой производилась оплата."},
		},
	}
	for _, c := range cases {
		t.Run(c.event, func(t *testing.T) {
			svc, _, mailer := newTestService()
			h := svc.Handlers()[c.event]
			if h == nil {
				t.Fatalf("нет обработчика %s", c.event)
			}
			if err := h(context.Background(), []byte(c.payload)); err != nil {
				t.Fatal(err)
			}
			if len(mailer.sent) != 1 {
				t.Fatalf("отправлено писем: %d", len(mailer.sent))
			}
			m := mailer.sent[0]
			if m.to != "anna@example.ru" || m.subject != c.subject {
				t.Fatalf("письмо: to=%q subject=%q", m.to, m.subject)
			}
			for _, s := range c.body {
				if !strings.Contains(m.body, s) {
					t.Errorf("в письме нет %q:\n%s", s, m.body)
				}
			}
			for _, s := range c.absent {
				if strings.Contains(m.body, s) {
					t.Errorf("в письме лишнее %q:\n%s", s, m.body)
				}
			}
		})
	}
}

func TestHandlersCoverAllEvents(t *testing.T) {
	svc, _, _ := newTestService()
	h := svc.Handlers()
	for _, e := range []string{
		notification.EventSubmissionApproved, notification.EventSubmissionRejected,
		notification.EventOrderPlaced, notification.EventOrderPaid,
		notification.EventOrderShipped, notification.EventOrderCancelled,
	} {
		if h[e] == nil {
			t.Errorf("нет обработчика %s", e)
		}
	}
	if len(h) != 6 {
		t.Errorf("обработчиков %d, ждали 6", len(h))
	}
}

func TestMissingRecipientSkipped(t *testing.T) {
	svc, _, mailer := newTestService()
	h := svc.Handlers()
	payloads := map[string]string{
		"удалённый пользователь":    `{"orderId":1,"userId":999,"totalKopecks":100}`,
		"покупатель удалён (null)":  `{"orderId":1,"userId":null,"totalKopecks":100}`,
		"у пользователя нет адреса": `{"orderId":1,"userId":8,"totalKopecks":100}`,
		"пользователь без userId":   `{"orderId":1,"totalKopecks":100}`,
	}
	for name, p := range payloads {
		if err := h[notification.EventOrderPaid](context.Background(), []byte(p)); err != nil {
			t.Errorf("%s: ожидался пропуск без ошибки, получили %v", name, err)
		}
	}
	if len(mailer.sent) != 0 {
		t.Fatalf("письма отправлены: %+v", mailer.sent)
	}
}

// Ошибка рода NotFound от модуля пользователей (например, auth.ErrUserNotFound) — тоже пропуск.
func TestAnyNotFoundErrorSkipped(t *testing.T) {
	svc, users, mailer := newTestService()
	users.err = errs.New(errs.NotFound, "пользователь не найден")
	err := svc.Handlers()[notification.EventSubmissionApproved](context.Background(),
		[]byte(`{"submissionId":1,"authorId":7,"title":"x","objectSlug":"y"}`))
	if err != nil || len(mailer.sent) != 0 {
		t.Fatalf("err=%v, писем %d", err, len(mailer.sent))
	}
}

func TestMalformedPayload(t *testing.T) {
	svc, _, mailer := newTestService()
	h := svc.Handlers()
	cases := map[string]struct{ event, payload string }{
		"не JSON":             {notification.EventOrderPaid, `{"orderId":`},
		"неверный тип поля":   {notification.EventOrderShipped, `{"orderId":"пятнадцать","userId":7}`},
		"нет orderId":         {notification.EventOrderCancelled, `{"userId":7}`},
		"нет автора заявки":   {notification.EventSubmissionApproved, `{"submissionId":1,"title":"x"}`},
		"нет номера заявки":   {notification.EventSubmissionRejected, `{"authorId":7,"reason":"x"}`},
		"нет номера в placed": {notification.EventOrderPlaced, `{"userId":7,"totalKopecks":100}`},
	}
	for name, c := range cases {
		err := h[c.event](context.Background(), []byte(c.payload))
		if !errors.Is(err, notification.ErrMalformedEvent) {
			t.Errorf("%s: ожидалась ErrMalformedEvent, получили %v", name, err)
		}
	}
	if len(mailer.sent) != 0 {
		t.Fatalf("письма отправлены: %+v", mailer.sent)
	}
}

// Ошибки почты и модуля пользователей возвращаются — outbox повторит событие.
func TestTransientErrorsReturned(t *testing.T) {
	payload := []byte(`{"orderId":1,"userId":7,"totalKopecks":100}`)

	svc, _, mailer := newTestService()
	mailer.err = errors.New("SMTP недоступен")
	if err := svc.Handlers()[notification.EventOrderPaid](context.Background(), payload); err == nil {
		t.Fatal("ошибка почты должна возвращаться")
	}

	svc, users, _ := newTestService()
	users.err = errors.New("БД недоступна")
	if err := svc.Handlers()[notification.EventOrderPaid](context.Background(), payload); err == nil {
		t.Fatal("ошибка поиска пользователя должна возвращаться")
	}
}

func TestHeaderInjectionRejected(t *testing.T) {
	svc, users, mailer := newTestService()
	users.byID[9] = contact{"evil@example.ru\r\nBcc: x@y.z", "Злоумышленник"}
	err := svc.Handlers()[notification.EventOrderPaid](context.Background(),
		[]byte(`{"orderId":1,"userId":9,"totalKopecks":100}`))
	if !errors.Is(err, notification.ErrHeaderInjection) || len(mailer.sent) != 0 {
		t.Fatalf("err=%v, писем %d", err, len(mailer.sent))
	}
}

func TestFormatRubles(t *testing.T) {
	for kopecks, want := range map[int64]string{450000: "4500.00", 195005: "1950.05", 5: "0.05", 0: "0.00", -150: "-1.50"} {
		if got := notification.FormatRubles(kopecks); got != want {
			t.Errorf("FormatRubles(%d) = %s, ждали %s", kopecks, got, want)
		}
	}
}
