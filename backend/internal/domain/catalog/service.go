package catalog

import "context"

// Service — сценарии модуля. Реализация — service/catalog; transport и другие модули
// зависят только от этого интерфейса.
type Service interface {
	ListCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, in CategoryInput) (Category, error)
	UpdateCategory(ctx context.Context, id int32, in CategoryInput) (Category, error)
	DeleteCategory(ctx context.Context, id int32) error

	// List — список объектов для посетителя v: неопубликованные попадают в него,
	// только если f.IncludeUnpublished и v — модератор.
	List(ctx context.Context, f Filter, v Viewer) ([]Object, error)
	// GetVisible — карточка объекта для посетителя v: неопубликованный объект
	// для гостя и обычного пользователя — ErrObjectNotFound.
	GetVisible(ctx context.Context, slug string, v Viewer) (Object, error)
	// GetBySlug — объект без проверки видимости, для других модулей (заявки).
	GetBySlug(ctx context.Context, slug string) (Object, error)
	// Create работает и в транзакции вызывающего модуля (одобрение заявки в TxManager.Do).
	Create(ctx context.Context, in ObjectInput) (Object, error)
	Update(ctx context.Context, id int64, in ObjectInput) (Object, error)
	Delete(ctx context.Context, id int64) error

	// Для других модулей.
	Exists(ctx context.Context, id int64) (bool, error)
	// PublishedPlaces — опубликованные объекты по slug, порядок не гарантируется.
	PublishedPlaces(ctx context.Context, slugs []string) ([]Place, error)
	// PlacesByIDs — объекты по id, порядок не гарантируется.
	PlacesByIDs(ctx context.Context, ids []int64) ([]Place, error)
}
