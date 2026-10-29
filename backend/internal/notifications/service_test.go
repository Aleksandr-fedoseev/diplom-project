package notifications

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"heritage-backend/internal/testdb"
)

// flakySender отказывает первые fails раз письмам на адрес failTo, остальные «отправляет» —
// запоминает. Адрес нужен потому, что в общей тестовой БД могут лежать письма других тестов.
type flakySender struct {
	failTo string
	fails  int
	sent   []Message
}

func (f *flakySender) Send(_ context.Context, msg Message) error {
	if msg.To == f.failTo && f.fails > 0 {
		f.fails--
		return errors.New("SMTP недоступен")
	}
	f.sent = append(f.sent, msg)
	return nil
}

func TestQueueRetriesAndSends(t *testing.T) {
	db := testdb.Connect(t)
	ctx := context.Background()
	to := testdb.Unique("reader") + "@example.ru"
	sender := &flakySender{failTo: to, fails: 1}
	svc := NewService(db, sender)

	if err := svc.Enqueue(ctx, Message{To: to, Subject: "Заявка одобрена", Body: "Спасибо!"}); err != nil {
		t.Fatal(err)
	}
	if err := svc.Enqueue(ctx, Message{To: to, Subject: "плохая\r\nBcc: x@y.z", Body: ""}); err == nil {
		t.Fatal("тема с переводом строки должна отклоняться")
	}

	// Первая попытка падает, письмо откладывается на минуту.
	if err := svc.ProcessDue(ctx); err != nil {
		t.Fatal(err)
	}
	var status, lastError string
	var attempts int
	var next time.Time
	err := db.QueryRow(ctx, `SELECT status, attempts, last_error, next_attempt_at FROM email_jobs WHERE to_email = $1`, to).
		Scan(&status, &attempts, &lastError, &next)
	if err != nil {
		t.Fatal(err)
	}
	if status != "pending" || attempts != 1 || lastError == "" || time.Until(next) < 30*time.Second {
		t.Fatalf("после ошибки: status=%s attempts=%d err=%q next=%v", status, attempts, lastError, next)
	}

	// Подошёл срок повтора — письмо уходит.
	if _, err := db.Exec(ctx, `UPDATE email_jobs SET next_attempt_at = now() WHERE to_email = $1`, to); err != nil {
		t.Fatal(err)
	}
	if err := svc.ProcessDue(ctx); err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, m := range sender.sent {
		found = found || (m.To == to && m.Subject == "Заявка одобрена")
	}
	if err := db.QueryRow(ctx, `SELECT status FROM email_jobs WHERE to_email = $1`, to).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if !found || status != "sent" {
		t.Fatalf("письмо не отправлено: status=%s, отправлено %d", status, len(sender.sent))
	}
}

func TestFailureGivesUpAfterMaxAttempts(t *testing.T) {
	if p := failure(1, 2, errors.New("x")); p.Status != "pending" || time.Until(p.NextAttemptAt) < 3*time.Minute {
		t.Fatalf("после 2-й попытки: %+v", p)
	}
	if p := failure(1, maxAttempts, errors.New("x")); p.Status != "failed" {
		t.Fatalf("после последней попытки: %+v", p)
	}
}

func TestBuildMessageEncodesCyrillic(t *testing.T) {
	msg := string(buildMessage("noreply@heritage.ru", Message{To: "a@b.ru", Subject: "Решение по заявке", Body: "Текст"}))
	if !strings.Contains(msg, "Subject: =?utf-8?q?") || !strings.Contains(msg, "0KLQtdC60YHRgg==") {
		t.Fatalf("письмо закодировано неверно:\n%s", msg)
	}
}
