package submission

import "context"

// Service — сценарии модуля. Реализация — service/submission; transport зависит только от этого интерфейса.
type Service interface {
	// Submit подаёт заявку. Файлы прикрепляются отдельными запросами (AddMedia) после создания.
	Submit(ctx context.Context, authorID int64, in NewSubmission) (Submission, error)
	// AddMedia прикрепляет файл к своей ожидающей заявке.
	AddMedia(ctx context.Context, submissionID, userID int64, in MediaInput) (File, error)
	// Get — карточка заявки с файлами и решением; модератору — ещё и сведения об авторе.
	Get(ctx context.Context, id int64, v Viewer) (Submission, error)
	// ListMine — свои заявки пользователя со статусом и причиной отказа, новые первыми.
	ListMine(ctx context.Context, authorID int64) ([]Submission, error)
	// Queue — очередь модерации: заявки с данным статусом, старые первыми.
	Queue(ctx context.Context, status Status, limit, offset int) ([]Submission, error)
	// Approve одобряет заявку: для новой — создаёт объект каталога, для дополнения —
	// передаёт файлы объекту. Всё в одной транзакции с событием submission.approved.
	Approve(ctx context.Context, id int64, mod Moderator, in ApproveInput) (Submission, error)
	// Reject отклоняет заявку с причиной (и, по желанию, блокирует автора)
	// в одной транзакции с событием submission.rejected.
	Reject(ctx context.Context, id int64, mod Moderator, in RejectInput) (Submission, error)
}
