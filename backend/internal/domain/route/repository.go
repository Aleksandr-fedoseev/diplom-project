package route

import "context"

// Repository — хранилище маршрутов. Реализация — repository/postgres.
// Маршруты без остановок: остановки хранятся отдельно (StopRef), а данные объектов — в каталоге.
type Repository interface {
	// Create сохраняет маршрут (без остановок) и возвращает его id. Повтор slug — ErrSlugTaken.
	Create(ctx context.Context, r Route) (int64, error)
	// AddStops сохраняет остановки маршрута (Position и ObjectID).
	AddStops(ctx context.Context, routeID int64, stops []Stop) error
	// BySlug возвращает маршрут с линией; нет такого — ErrNotFound.
	BySlug(ctx context.Context, slug string) (Route, error)
	// Списки — без линий.
	ListThematic(ctx context.Context) ([]Route, error)
	ListByOwner(ctx context.Context, ownerID int64) ([]Route, error)
	// StopRefs — точки сразу нескольких маршрутов, по маршруту и номеру остановки.
	StopRefs(ctx context.Context, routeIDs []int64) ([]StopRef, error)
	// Owner — владелец маршрута (nil у тематического); нет такого — ErrNotFound.
	Owner(ctx context.Context, id int64) (*int64, error)
	Delete(ctx context.Context, id int64) error
}
