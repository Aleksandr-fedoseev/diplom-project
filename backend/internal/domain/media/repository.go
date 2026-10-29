package media

import "context"

// Repository — сведения о файлах в БД. Реализация — repository/postgres.
// ByID и Delete возвращают ErrFileNotFound, если файла нет.
// Методы работают и в транзакции вызывающего (соединение берётся из ctx).
type Repository interface {
	Create(ctx context.Context, f File) (File, error)
	ByID(ctx context.Context, id int64) (File, error)
	ListByObject(ctx context.Context, objectID int64) ([]File, error)
	ListBySubmission(ctx context.Context, submissionID int64) ([]File, error)
	// AttachToObject передаёт объекту все файлы заявки; связь с заявкой остаётся для истории.
	AttachToObject(ctx context.Context, submissionID, objectID int64) error
	Delete(ctx context.Context, id int64) error
}
