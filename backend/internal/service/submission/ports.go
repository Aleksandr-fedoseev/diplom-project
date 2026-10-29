package submission

import (
	"context"
	"io"

	"heritage-backend/internal/domain/submission"
)

// Порты сервиса: всё, что заявкам нужно от других модулей и техники. Типы здесь свои —
// модуль не импортирует чужие пакеты, а адаптеры при сборке приложения переводят
// вызовы в сервисы каталога, медиа и пользователей. В тестах подставляются фейки.

// TxManager выполняет fn в транзакции: транзакция передаётся через ctx, и все
// репозитории и порты, вызванные с этим ctx, работают в ней. Ошибка fn — откат.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

// EventPublisher записывает событие в outbox в текущей транзакции (из ctx).
type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload any) error // payload кодируется в JSON
}

// ObjectRef — краткие сведения об объекте каталога.
type ObjectRef struct {
	ID        int64
	Slug      string
	Title     string
	Published bool
}

// NewObject — объект каталога, создаваемый при одобрении новой заявки.
type NewObject struct {
	Slug               string
	Title              string
	Category           string
	ProtectionStatus   string
	RegistryNumber     *string
	YearBuilt          *string
	Architect          *string
	ArchitecturalStyle *string
	Address            string
	Lat                float64
	Lng                float64
	ShortDescription   string
	FullHistory        string
	Tags               []string
	IsPublished        bool
}

// Catalog — модуль каталога объектов.
type Catalog interface {
	// ObjectBySlug — объект по slug. Нет объекта или он не опубликован — submission.ErrObjectNotFound.
	ObjectBySlug(ctx context.Context, slug string) (ObjectRef, error)
	// CreateObject создаёт объект (в транзакции из ctx). Неверные поля — ошибка домена рода Invalid,
	// занятый slug — Conflict.
	CreateObject(ctx context.Context, in NewObject) (ObjectRef, error)
	// ObjectsByIDs — объекты по списку id (для objectSlug/objectTitle в списках); несуществующие пропускаются.
	ObjectsByIDs(ctx context.Context, ids []int64) ([]ObjectRef, error)
}

// MediaFile — файл заявки в модуле медиа (тот же тип, что в карточке заявки).
type MediaFile = submission.File

// MediaUpload — файл, который нужно сохранить и привязать к заявке.
type MediaUpload struct {
	SubmissionID int64
	Kind         string
	Caption      string
	YearTaken    string
	SourceNote   string
	OriginalName string
	Size         int64
	Body         io.Reader
	UploadedBy   int64
}

// Media — модуль медиафайлов.
type Media interface {
	Upload(ctx context.Context, in MediaUpload) (MediaFile, error)
	ListBySubmission(ctx context.Context, submissionID int64) ([]MediaFile, error)
	// AttachToObject передаёт все файлы заявки объекту (в транзакции из ctx).
	AttachToObject(ctx context.Context, submissionID, objectID int64) error
}

// Author — пользователь, как его видит модуль заявок (тот же тип, что в карточке заявки).
type Author = submission.Author

// Users — модуль пользователей.
type Users interface {
	User(ctx context.Context, id int64) (Author, error)
	// EnsureActive — ошибка рода Forbidden, если пользователь заблокирован.
	EnsureActive(ctx context.Context, id int64) error
	// Block блокирует пользователя от имени модератора (в транзакции из ctx). Права проверяет
	// модуль пользователей: модератор не может заблокировать другого модератора.
	Block(ctx context.Context, actorID int64, actorRole string, userID int64) error
}
