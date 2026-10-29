package catalog

import "context"

// ObjectRepository — хранилище объектов. Реализация — repository/postgres.
// Методы возвращают ошибки домена: ErrObjectNotFound, ErrSlugTaken, ErrUnknownCategory, ErrObjectInRoute.
// Все методы работают и внутри транзакции вызывающего (TxManager.Do): соединение берётся из ctx.
type ObjectRepository interface {
	List(ctx context.Context, f Filter) ([]Object, error)
	BySlug(ctx context.Context, slug string) (Object, error)
	// Create сохраняет объект и возвращает его, перечитав через то же соединение.
	Create(ctx context.Context, in ObjectInput) (Object, error)
	Update(ctx context.Context, id int64, in ObjectInput) (Object, error)
	Delete(ctx context.Context, id int64) error
	Exists(ctx context.Context, id int64) (bool, error)
	PublishedPlaces(ctx context.Context, slugs []string) ([]Place, error)
	PlacesByIDs(ctx context.Context, ids []int64) ([]Place, error)
}

// CategoryRepository — хранилище категорий.
// Ошибки: ErrCategoryNotFound, ErrCategorySlugTaken, ErrCategoryNotEmpty.
type CategoryRepository interface {
	List(ctx context.Context) ([]Category, error)
	Create(ctx context.Context, in CategoryInput) (Category, error)
	Update(ctx context.Context, id int32, in CategoryInput) (Category, error)
	Delete(ctx context.Context, id int32) error
}
