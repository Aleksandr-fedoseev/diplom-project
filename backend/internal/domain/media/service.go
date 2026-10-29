package media

import (
	"context"
	"io"
)

// Service — сценарии модуля. Реализация — service/media; transport и другие модули
// зависят только от этого интерфейса.
type Service interface {
	// Upload сохраняет файл к объекту или к заявке (ровно одно из in.ObjectID, in.SubmissionID).
	Upload(ctx context.Context, in UploadInput) (File, error)
	ListByObject(ctx context.Context, objectID int64) ([]File, error)
	ListBySubmission(ctx context.Context, submissionID int64) ([]File, error)
	// Open возвращает файл и поток его содержимого. Поток умеет Seek, поэтому HTTP может
	// отдавать файл частями (Range). Файл, который v видеть не положено, — ErrFileNotFound.
	// Поток закрывает вызывающий.
	Open(ctx context.Context, id int64, v Viewer) (File, io.ReadSeekCloser, error)
	Delete(ctx context.Context, id int64) error
	// AttachToObject передаёт файлы одобренной заявки объекту; вызывается в транзакции модуля заявок.
	AttachToObject(ctx context.Context, submissionID, objectID int64) error
	// MaxUploadBytes — общий предел размера загружаемого файла.
	MaxUploadBytes() int64
}
