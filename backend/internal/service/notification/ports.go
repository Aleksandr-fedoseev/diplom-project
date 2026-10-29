package notification

import "context"

// Порты сервиса. Реализации подставляются при сборке приложения, в тестах — фейки.

// Users находит контакт получателя письма. Реализация — адаптер к модулю auth.
// Если пользователя нет, Contact возвращает ошибку рода errs.NotFound
// (например, notification.ErrRecipientNotFound или auth.ErrUserNotFound) — письмо тогда пропускается.
type Users interface {
	Contact(ctx context.Context, userID int64) (email, name string, err error)
}

// Mailer отправляет одно письмо. Реализации — infrastructure/smtp (Mailer, LogMailer).
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}
