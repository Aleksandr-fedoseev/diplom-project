// Package notifications — email-уведомления через очередь задач в БД.
// Модули ставят письмо в очередь (Enqueue) и сразу отвечают пользователю,
// а фоновый обработчик (Run) отправляет письма и при ошибке повторяет попытку.
package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/notifications/store"
)

const (
	maxAttempts  = 5
	batchSize    = 20
	pollInterval = 5 * time.Second
)

type Message struct {
	To      string
	Subject string
	Body    string
}

// Sender отправляет одно письмо. Реализации: SMTPSender и LogSender.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type Service struct {
	q      *store.Queries // запросы из queries.sql, код сгенерирован sqlc
	sender Sender
}

func NewService(db *pgxpool.Pool, sender Sender) *Service {
	return &Service{q: store.New(db), sender: sender}
}

// Enqueue ставит письмо в очередь.
func (s *Service) Enqueue(ctx context.Context, msg Message) error {
	return enqueue(ctx, s.q, msg)
}

// EnqueueTx ставит письмо в очередь в транзакции вызывающего модуля:
// письмо уйдёт, только если вся операция (например, решение по заявке) сохранилась.
func (s *Service) EnqueueTx(ctx context.Context, tx pgx.Tx, msg Message) error {
	return enqueue(ctx, s.q.WithTx(tx), msg)
}

func enqueue(ctx context.Context, q *store.Queries, msg Message) error {
	// Перевод строки в адресе или теме позволил бы дописать в письмо чужие заголовки.
	if strings.ContainsAny(msg.To+msg.Subject, "\r\n") {
		return fmt.Errorf("перевод строки в адресе или теме письма")
	}
	err := q.EnqueueEmail(ctx, store.EnqueueEmailParams{ToEmail: msg.To, Subject: msg.Subject, Body: msg.Body})
	if err != nil {
		return fmt.Errorf("постановка письма в очередь: %w", err)
	}
	return nil
}

// Run обрабатывает очередь, пока не отменён ctx. Запускается в отдельной горутине из main.
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		if err := s.ProcessDue(ctx); err != nil && ctx.Err() == nil {
			slog.Error("обработка очереди писем", "err", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// ProcessDue отправляет письма, срок которых наступил. Вынесено из Run, чтобы вызывать в тестах.
func (s *Service) ProcessDue(ctx context.Context) error {
	jobs, err := s.q.ClaimDueEmails(ctx, batchSize)
	if err != nil {
		return err
	}
	// UPDATE ... RETURNING не сохраняет порядок — отправляем в порядке постановки в очередь,
	// чтобы «заказ оформлен» не пришёл после «заказ отменён».
	slices.SortFunc(jobs, func(a, b store.ClaimDueEmailsRow) int { return int(a.ID - b.ID) })
	for _, job := range jobs {
		sendErr := s.sender.Send(ctx, Message{To: job.ToEmail, Subject: job.Subject, Body: job.Body})
		if sendErr == nil {
			err = s.q.MarkEmailSent(ctx, job.ID)
		} else {
			err = s.q.MarkEmailFailed(ctx, failure(job.ID, job.Attempts, sendErr))
			slog.Warn("письмо не отправлено", "id", job.ID, "attempt", job.Attempts, "err", sendErr)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// failure решает, что делать с неотправленным письмом: повторить позже
// (через 1, 4, 9, 16 минут) или после последней попытки пометить как failed.
func failure(id int64, attempts int32, sendErr error) store.MarkEmailFailedParams {
	p := store.MarkEmailFailedParams{
		ID:            id,
		Status:        "pending",
		LastError:     sendErr.Error(),
		NextAttemptAt: time.Now().Add(time.Duration(attempts*attempts) * time.Minute),
	}
	if attempts >= maxAttempts {
		p.Status = "failed"
	}
	return p
}
