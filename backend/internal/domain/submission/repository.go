package submission

import "context"

// Repository — хранилище заявок и решений. Реализация — repository/postgres.
// Возвращает ошибки домена: ErrNotFound, если заявки нет; ErrAlreadyDecided при втором решении.
// Поля ObjectSlug, ObjectTitle, Author и Media репозиторий не заполняет — это делает сервис.
type Repository interface {
	// Create сохраняет новую заявку (статус pending) и возвращает её id.
	Create(ctx context.Context, s Submission) (int64, error)
	// ByID — заявка вместе с решением модератора, если оно уже есть.
	ByID(ctx context.Context, id int64) (Submission, error)
	// LockForDecision читает заявку и блокирует её строку до конца транзакции
	// (SELECT … FOR UPDATE). Вызывается только внутри TxManager.Do: второй модератор,
	// решающий ту же заявку, ждёт окончания первой транзакции и видит новый статус.
	LockForDecision(ctx context.Context, id int64) (Submission, error)
	// ListByAuthor — заявки автора, новые первыми.
	ListByAuthor(ctx context.Context, authorID int64) ([]Submission, error)
	// ListByStatus — очередь модерации, старые заявки первыми.
	ListByStatus(ctx context.Context, status Status, limit, offset int) ([]Submission, error)
	// SetStatus меняет статус; objectID != nil — заодно запоминает объект (созданный при одобрении).
	SetStatus(ctx context.Context, id int64, status Status, objectID *int64) error
	// CreateDecision сохраняет решение; по заявке оно одно (иначе ErrAlreadyDecided).
	CreateDecision(ctx context.Context, submissionID int64, d Decision) error
}
